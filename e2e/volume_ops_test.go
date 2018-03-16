package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"testing"

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

	tmpDir, err = ioutil.TempDir("", t.Name())
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
					{NodeID: gds[0].PeerID(), Path: brickPaths[0]},
					{NodeID: gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{NodeID: gds[0].PeerID(), Path: brickPaths[2]},
					{NodeID: gds[1].PeerID(), Path: brickPaths[3]},
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
			{NodeID: gds[0].PeerID(), Path: brickPaths[0]},
			{NodeID: gds[1].PeerID(), Path: brickPaths[1]},
			{NodeID: gds[0].PeerID(), Path: brickPaths[2]},
			{NodeID: gds[1].PeerID(), Path: brickPaths[3]},
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
	_, err := exec.Command("sh", "-c", "which glusterfsd").Output()
	if err != nil {
		t.SkipNow()
	}

	r := require.New(t)

	gds, err := setupCluster("./config/1.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", t.Name())
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
					{NodeID: gds[0].PeerID(), Path: brickPath},
				},
			},
		},
		Force: true,
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
		Name: "profile.test2",
		Options: []api.VolumeOption{{"opt1", "on", "off"},
			{"opt2", "enable", "disable"},
			{"opt3", "off", "on"}}}
	err = client.OptionGroupCreate(optionGroupReq)
	r.NotNil(err)

	optionGroupReq = api.OptionGroupReq{
		Name: "profile.test2",
		Options: []api.VolumeOption{{"afr.eager-lock", "on", "off"},
			{"gfproxy.afr.eager-lock", "on", "off"}}}
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
					{NodeID: gds[0].PeerID(), Path: brickPaths[0]},
					{NodeID: gds[1].PeerID(), Path: brickPaths[1]},
					{NodeID: gds[0].PeerID(), Path: brickPaths[2]},
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

	host, _, _ := net.SplitHostPort(gds[0].ClientAddress)

	mntCmd := exec.Command("mount", "-t", "glusterfs", host+":"+disperseVolName, mntPath)

	umntCmd := exec.Command("umount", mntPath)

	err = mntCmd.Run()
	r.Nil(err, fmt.Sprintf("disperse volume mount failed: %s", err))

	err = umntCmd.Run()
	r.Nil(err, fmt.Sprintf("disperse volume unmount failed: %s", err))

	r.Nil(client.VolumeStop(disperseVolName), "disperse volume stop failed")
	r.Nil(client.VolumeDelete(disperseVolName), "disperse volume delete failed")
}
