package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

const (
	volname = "testvol"
)

var (
	gds    []*gdProcess
	client *restclient.Client
	tmpDir string
)

// TestVolume creates a volume and starts it, runs further tests on it and
// finally deletes the volume
func TestVolume(t *testing.T) {
	var err error

	r := require.New(t)

	gds, err = setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	client = initRestclient(gds[0].ClientAddress)

	tmpDir, err = ioutil.TempDir(baseWorkdir, t.Name())
	r.Nil(err)
	t.Logf("Using temp dir: %s", tmpDir)
	//defer os.RemoveAll(tmpDir)

	// Create the volume
	t.Run("Create", testVolumeCreate)

	// Expand the volume
	t.Run("Expand", testVolumeExpand)

	// Run tests that depend on this volume
	t.Run("Start", testVolumeStart)
	t.Run("Mount", testVolumeMount)
	t.Run("Status", testVolumeStatus)
	t.Run("Statedump", testVolumeStatedump)
	t.Run("Stop", testVolumeStop)
	t.Run("List", testVolumeList)
	t.Run("Info", testVolumeInfo)

	// delete volume
	t.Run("Delete", testVolumeDelete)

	// Disperse volume test
	t.Run("Disperse", testDisperse)
}

func testVolumeCreate(t *testing.T) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Force: true,
	}
	_, err := client.VolumeCreate(createReq)
	r.Nil(err)
}

func testVolumeExpand(t *testing.T) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	expandReq := api.VolExpandReq{
		Bricks: []api.BrickReq{
			{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
			{PeerID: gds[1].PeerID(), Path: brickPaths[1]},
			{PeerID: gds[0].PeerID(), Path: brickPaths[2]},
			{PeerID: gds[1].PeerID(), Path: brickPaths[3]},
		},
		Force: true,
	}
	_, err := client.VolumeExpand(volname, expandReq)
	r.Nil(err)
}

func testVolumeDelete(t *testing.T) {
	r := require.New(t)
	r.Nil(client.VolumeDelete(volname))
}

func testVolumeStart(t *testing.T) {
	r := require.New(t)

	r.Nil(client.VolumeStart(volname), "volume start failed")
}

func testVolumeStop(t *testing.T) {
	r := require.New(t)

	r.Nil(client.VolumeStop(volname), "volume stop failed")
}

func testVolumeList(t *testing.T) {
	r := require.New(t)

	volumes, err := client.Volumes("")
	r.Nil(err)
	r.Len(volumes, 1)
}

func testVolumeInfo(t *testing.T) {
	r := require.New(t)

	_, err := client.Volumes(volname)
	r.Nil(err)
}

func testVolumeStatus(t *testing.T) {
	r := require.New(t)

	_, err := client.VolumeStatus(volname)
	r.Nil(err)
}

func testVolumeStatedump(t *testing.T) {
	r := require.New(t)

	// Get statedump dir
	var statedumpDir string
	args := []string{"--print-statedumpdir"}
	cmdOut, err := exec.Command("glusterfsd", args...).Output()
	if err == nil {
		statedumpDir = strings.TrimSpace(string(cmdOut))
	} else {
		// fallback to hard-coded value
		statedumpDir = "/var/run/gluster"
	}

	// statedump file pattern: hyphenated-brickpath.<pid>.dump.<timestamp>
	pattern := statedumpDir + "/*[0-9]*.dump.[0-9]*"

	// remove old statedump files
	files, err := filepath.Glob(pattern)
	r.Nil(err)
	for _, f := range files {
		os.Remove(f)
	}

	// take statedump
	var req api.VolStatedumpReq
	req.Bricks = true
	r.Nil(client.VolumeStatedump(volname, req))
	// give it some time to ensure the statedumps are generated
	time.Sleep(1 * time.Second)

	// Check if statedump have been generated for all bricks
	files, err = filepath.Glob(pattern)
	r.Nil(err)
	r.Equal(len(files), 8) // 4 bricks during vol create + 4 after expand
}

// testVolumeMount mounts checks if the volume mounts successfully and unmounts it
func testVolumeMount(t *testing.T) {
	r := require.New(t)

	mntPath, err := ioutil.TempDir(tmpDir, "mnt")
	r.Nil(err)
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(gds[0].ClientAddress)
	mntCmd := exec.Command("mount", "-t", "glusterfs", host+":"+volname, mntPath)
	umntCmd := exec.Command("umount", mntPath)

	err = mntCmd.Run()
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	err = umntCmd.Run()
	r.Nil(err, fmt.Sprintf("unmount failed: %s", err))
}

func TestVolumeOptions(t *testing.T) {

	// skip this test if glusterfs server packages and xlators are not
	// installed
	_, err := exec.Command("sh", "-c", "command -v glusterfsd").Output()
	if err != nil {
		t.SkipNow()
	}

	r := require.New(t)

	gds, err := setupCluster("./config/1.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir(baseWorkdir, t.Name())
	defer os.RemoveAll(brickDir)

	brickPath, err := ioutil.TempDir(brickDir, "brick")
	r.Nil(err)

	client := initRestclient(gds[0].ClientAddress)

	volname := "testvol"
	createReq := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPath},
				},
			},
		},
		Force: true,
		// XXX: Setting advanced, as all options are advanced by default
		// TODO: Remove this later if the default changes
		Advanced: true,
	}

	// valid option test cases
	validOpKeys := []string{"gfproxy.afr.eager-lock", "afr.eager-lock"}
	for _, validKey := range validOpKeys {
		createReq.Options = map[string]string{validKey: "on"}

		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		err = client.VolumeDelete(volname)
		r.Nil(err)
	}

	// invalid option test cases
	invalidOpKeys := []string{"..eager-lock", "a.b.afr.eager-lock", "afr.non-existent", "eager-lock"}
	for _, invalidKey := range invalidOpKeys {
		createReq.Options = map[string]string{invalidKey: "on"}

		_, err = client.VolumeCreate(createReq)
		r.NotNil(err)
	}

	// test options that are settable and not settable
	createReq.Options = nil
	_, err = client.VolumeCreate(createReq)
	r.Nil(err)
	var optionReq api.VolOptionReq
	// XXX: Setting advanced, as all options are advanced by default
	// TODO: Remove this later if the default changes
	optionReq.Advanced = true

	settableKey := "afr.use-compound-fops"
	optionReq.Options = map[string]string{settableKey: "on"}
	r.Nil(client.VolumeSet(volname, optionReq))
	notSettableKey := "afr.consistent-io"
	optionReq.Options = map[string]string{notSettableKey: "on"}
	r.NotNil(client.VolumeSet(volname, optionReq))
	r.Nil(client.VolumeDelete(volname))

	// group option test cases
	groupOpKeys := []string{"profile.test"}
	for _, validKey := range groupOpKeys {
		createReq.Options = map[string]string{validKey: "on"}

		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		err = client.VolumeDelete(volname)
		r.Nil(err)
	}

	for _, validKey := range groupOpKeys {
		createReq.Options = map[string]string{validKey: "off"}

		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		err = client.VolumeDelete(volname)
		r.Nil(err)
	}

	optionGroupReq := api.OptionGroupReq{
		OptionGroup: api.OptionGroup{
			Name: "profile.test2",
			Options: []api.VolumeOption{{Name: "opt1", OnValue: "on"},
				{Name: "opt2", OnValue: "enable"},
				{Name: "opt3", OnValue: "off"}},
			Description: "Test profile 2",
		},
		// XXX: Setting advanced, as all options are advanced by default
		// TODO: Remove this later if the default changes
		Advanced: true,
	}
	err = client.OptionGroupCreate(optionGroupReq)
	r.NotNil(err)

	optionGroupReq = api.OptionGroupReq{
		OptionGroup: api.OptionGroup{
			Name: "profile.test2",
			Options: []api.VolumeOption{{Name: "afr.eager-lock", OnValue: "on"},
				{Name: "gfproxy.afr.eager-lock", OnValue: "on"},
			},
			Description: "Test profile 2",
		},
		// XXX: Setting advanced, as all options are advanced by default
		// TODO: Remove this later if the default changes
		Advanced: true,
	}
	err = client.OptionGroupCreate(optionGroupReq)
	r.Nil(err)

	_, err = client.OptionGroupList()
	r.Nil(err)

	r.Nil(client.OptionGroupDelete("profile.test2"))

}

func testDisperse(t *testing.T) {
	r := require.New(t)
	disperseVolName := "dispersetestvol"

	var brickPaths []string
	for i := 1; i <= 3; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	createReq := api.VolCreateReq{
		Name: disperseVolName,
		Subvols: []api.SubvolReq{
			{
				Type: "disperse",
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[1]},
					{PeerID: gds[0].PeerID(), Path: brickPaths[2]},
				},
				DisperseRedundancy: 1,
			},
		},
		Force: true,
	}

	_, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeStart(disperseVolName), "disperse volume start failed")

	mntPath, err := ioutil.TempDir(tmpDir, "mnt")
	r.Nil(err)
	defer os.RemoveAll(mntPath)

	time.Sleep(1 * time.Second)

	host, _, _ := net.SplitHostPort(gds[0].ClientAddress)

	mntCmd := exec.Command("mount", "-t", "glusterfs", host+":"+disperseVolName, mntPath)
	err = mntCmd.Run()
	r.Nil(err, fmt.Sprintf("disperse volume mount failed: %s", err))

	umntCmd := exec.Command("umount", mntPath)
	err = umntCmd.Run()
	r.Nil(err, fmt.Sprintf("disperse volume unmount failed: %s", err))

	r.Nil(client.VolumeStop(disperseVolName), "disperse volume stop failed")
	r.Nil(client.VolumeDelete(disperseVolName), "disperse volume delete failed")
}
