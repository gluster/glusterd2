package e2e

import (
	"errors"
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
	// for the disperse volume tests
	disperseVolName = "dispersetestvol"
)

var (
	client *restclient.Client
)

// TestVolume creates a volume and starts it, runs further tests on it and
// finally deletes the volume
func TestVolume(t *testing.T) {
	var err error

	r := require.New(t)

	tc, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client = initRestclient(tc.gds[0])

	t.Run("CreateWithoutName", tc.wrap(testVolumeCreateWithoutName))

	// Create the volume
	t.Run("Create", tc.wrap(testVolumeCreate))
	// Expand the volume
	t.Run("Expand", tc.wrap(testVolumeExpand))

	// Run tests that depend on this volume
	t.Run("Start", testVolumeStart)
	t.Run("Mount", tc.wrap(testVolumeMount))
	t.Run("Status", testVolumeStatus)
	t.Run("Statedump", testVolumeStatedump)
	t.Run("Stop", testVolumeStop)
	t.Run("List", testVolumeList)
	t.Run("Info", testVolumeInfo)
	t.Run("Edit", testEditVolume)
	t.Run("VolumeFlags", tc.wrap(testVolumeCreateWithFlags))
	// delete volume
	t.Run("Delete", testVolumeDelete)

	// Disperse volume test
	t.Run("Disperse", tc.wrap(testDisperse))
	t.Run("DisperseMount", tc.wrap(testDisperseMount))
	t.Run("DisperseDelete", testDisperseDelete)
}

func testVolumeCreateWithoutName(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 2; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
		Subvols: []api.SubvolReq{
			{
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeDelete(volinfo.Name))
}

func testVolumeCreate(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath := testTempDir(t, "brick")
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
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Metadata: map[string]string{
			"owner": "gd2test",
		},
		Force: true,
	}
	_, err := client.VolumeCreate(createReq)
	r.Nil(err)

	//invalid volume name
	createReq.Name = "##@@#@!#@!!@#"
	_, err = client.VolumeCreate(createReq)
	r.NotNil(err)

	testDisallowBrickReuse(t, brickPaths[0], tc)
}

func testDisallowBrickReuse(t *testing.T, brickInUse string, tc *testCluster) {
	r := require.New(t)
	volname := formatVolName(t.Name())

	createReq := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickInUse},
				},
			},
		},
		Force: true,
	}

	_, err := client.VolumeCreate(createReq)
	r.NotNil(err)
}

func testVolumeCreateWithFlags(t *testing.T, tc *testCluster) {
	r := require.New(t)
	volumeName := formatVolName(t.Name())
	var brickPaths []string

	for i := 1; i <= 4; i++ {
		brickPaths = append(brickPaths, fmt.Sprintf(baseLocalStateDir+"/"+t.Name()+"/%d", i))
	}

	flags := make(map[string]bool)
	//set flags to allow rootdir
	flags["allow-root-dir"] = true
	//set flags create brick dir
	flags["create-brick-dir"] = true

	createReqBrick := api.VolCreateReq{
		Name: volumeName,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Flags: flags,
	}

	_, err := client.VolumeCreate(createReqBrick)
	r.Nil(err)

	//delete volume
	r.Nil(client.VolumeDelete(volumeName))

	createReqBrick.Name = volumeName
	//set reuse-brick flag
	flags["reuse-bricks"] = true
	createReqBrick.Flags = flags

	_, err = client.VolumeCreate(createReqBrick)
	r.Nil(err)

	r.Nil(client.VolumeDelete(volumeName))

	//recreate deleted volume
	_, err = client.VolumeCreate(createReqBrick)
	r.Nil(err)

	//delete volume
	r.Nil(client.VolumeDelete(volumeName))

}

func testVolumeExpand(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPaths = append(brickPaths, fmt.Sprintf(fmt.Sprintf(baseLocalStateDir+"/"+t.Name()+"/%d/", i)))
	}

	flags := make(map[string]bool)
	//set flags to allow rootdir and create brick dir
	flags["create-brick-dir"] = true
	flags["allow-root-dir"] = true

	expandReq := api.VolExpandReq{
		Bricks: []api.BrickReq{
			{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
			{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
			{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
			{PeerID: tc.gds[1].PeerID(), Path: brickPaths[3]},
		},
		Flags: flags,
	}

	//expand with new brick dir which is not created
	volinfo, err := client.VolumeExpand(volname, expandReq)
	r.Nil(err)

	// Two subvolumes are added to the volume created by testVolumeCreate,
	// total expected subvols is 4. Each subvol should contain two bricks
	// since Volume type is Replica
	r.Len(volinfo.Subvols, 4)
	for _, subvol := range volinfo.Subvols {
		r.Len(subvol.Bricks, 2)
	}
}

func testVolumeDelete(t *testing.T) {
	r := require.New(t)
	r.Nil(client.VolumeDelete(volname))
}

func testVolumeStart(t *testing.T) {
	r := require.New(t)
	r.Nil(client.VolumeStart(volname, false), "volume start failed")
}

func testVolumeStop(t *testing.T) {
	r := require.New(t)

	r.Nil(client.VolumeStop(volname), "volume stop failed")
}

func testVolumeList(t *testing.T) {
	r := require.New(t)
	var matchingQueries []map[string]string
	var nonMatchingQueries []map[string]string

	matchingQueries = append(matchingQueries, map[string]string{
		"key":   "owner",
		"value": "gd2test",
	})
	matchingQueries = append(matchingQueries, map[string]string{
		"key": "owner",
	})
	matchingQueries = append(matchingQueries, map[string]string{
		"value": "gd2test",
	})
	for _, filter := range matchingQueries {
		volumes, err := client.Volumes("", filter)
		r.Nil(err)
		r.Len(volumes, 1)
	}

	nonMatchingQueries = append(nonMatchingQueries, map[string]string{
		"key":   "owner",
		"value": "gd2-test",
	})
	nonMatchingQueries = append(nonMatchingQueries, map[string]string{
		"key": "owners",
	})
	nonMatchingQueries = append(nonMatchingQueries, map[string]string{
		"value": "gd2tests",
	})
	for _, filter := range nonMatchingQueries {
		volumes, err := client.Volumes("", filter)
		r.Nil(err)
		r.Len(volumes, 0)
	}

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
	if _, err := os.Lstat("/dev/fuse"); os.IsNotExist(err) {
		t.Skip("skipping mount /dev/fuse unavailable")
	}
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
func testVolumeMount(t *testing.T, tc *testCluster) {
	testMountUnmount(t, volname, tc)
}

func testMountUnmount(t *testing.T, v string, tc *testCluster) {
	if _, err := os.Lstat("/dev/fuse"); os.IsNotExist(err) {
		t.Skip("skipping mount /dev/fuse unavailable")
	}
	r := require.New(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)
	mntCmd := exec.Command("mount", "-t", "glusterfs", host+":"+v, mntPath)
	umntCmd := exec.Command("umount", mntPath)

	err := mntCmd.Run()
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

	tc, err := setupCluster("./config/1.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	brickDir, err := ioutil.TempDir(baseLocalStateDir, t.Name())
	defer os.RemoveAll(brickDir)

	brickPath, err := ioutil.TempDir(brickDir, "brick")
	r.Nil(err)

	client := initRestclient(tc.gds[0])

	volname := "testvol"
	createReq := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPath},
				},
			},
		},
		Force: true,
		// XXX: Setting advanced, as all options are advanced by default
		// TODO: Remove this later if the default changes
		Advanced: true,
	}

	validOpKeys := []string{"gfproxy.afr.eager-lock", "afr.eager-lock"}
	invalidOpKeys := []string{"..eager-lock", "a.b.afr.eager-lock", "afr.non-existent", "eager-lock"}

	// valid option test cases
	for _, validKey := range validOpKeys {
		createReq.Options = map[string]string{validKey: "on"}

		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		// test volume get on valid keys
		_, err = client.VolumeGet(volname, validKey)
		r.Nil(err)

		var resetOptionReq api.VolOptionResetReq
		resetOptionReq.Options = append(resetOptionReq.Options, validKey)
		resetOptionReq.Force = true
		r.Nil(client.VolumeReset(volname, resetOptionReq))

		err = client.VolumeDelete(volname)
		r.Nil(err)
	}

	// invalid option test cases
	for _, invalidKey := range invalidOpKeys {
		createReq.Options = map[string]string{}
		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		_, err = client.VolumeGet(volname, invalidKey)
		r.NotNil(err)

		err = client.VolumeDelete(volname)
		r.Nil(err)

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

	var resetOptionReq api.VolOptionResetReq
	resetOptionReq.Options = []string{"afr.use-compound-fops"}
	resetOptionReq.Force = true
	r.Nil(client.VolumeReset(volname, resetOptionReq))

	validOpKeys = []string{"io-stats.count-fop-hits", "io-stats.latency-measurement"}
	for _, validKey := range validOpKeys {
		optionReq.Options = map[string]string{validKey: "on"}
		r.Nil(client.VolumeSet(volname, optionReq))
	}

	resetOptionReq.Force = true
	resetOptionReq.All = true
	r.Nil(client.VolumeReset(volname, resetOptionReq))

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

func testDisperse(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 3; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	createReq := api.VolCreateReq{
		Name: disperseVolName,
		Subvols: []api.SubvolReq{
			{
				Type: "disperse",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
				},
				DisperseRedundancy: 1,
			},
		},
		Force: true,
	}

	_, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeStart(disperseVolName, true), "disperse volume start failed")
}

func testDisperseMount(t *testing.T, tc *testCluster) {
	testMountUnmount(t, disperseVolName, tc)
}

func testDisperseDelete(t *testing.T) {
	r := require.New(t)
	r.Nil(client.VolumeStop(disperseVolName), "disperse volume stop failed")
	r.Nil(client.VolumeDelete(disperseVolName), "disperse volume delete failed")
}

func validateVolumeEdit(volinfo api.VolumeGetResp, editMetadataReq api.VolEditReq, resp api.VolumeEditResp) error {
	if editMetadataReq.DeleteMetadata {
		for key := range editMetadataReq.Metadata {
			_, existinVolinfo := volinfo.Metadata[key]
			_, existinResp := resp.Metadata[key]
			if existinVolinfo || existinResp {
				err := errors.New("invalid response")
				return err
			}
		}
	} else {
		for key, value := range editMetadataReq.Metadata {
			if volinfo.Metadata[key] != value || resp.Metadata[key] != value {
				err := errors.New("invalid response")
				return err
			}
		}
	}
	return nil
}

func testEditVolume(t *testing.T) {
	r := require.New(t)
	editMetadataReq := api.VolEditReq{
		Metadata: map[string]string{
			"owner": "gd2tests",
		},
		DeleteMetadata: false,
	}
	resp, err := client.EditVolume(volname, editMetadataReq)
	r.Nil(err)
	volinfo, err := client.Volumes(volname)
	r.Nil(err)
	err = validateVolumeEdit(volinfo[0], editMetadataReq, resp)
	r.Nil(err)
	editMetadataReq = api.VolEditReq{
		Metadata: map[string]string{
			"owner": "gd2functests",
			"year":  "2018",
		},
		DeleteMetadata: false,
	}
	resp, err = client.EditVolume(volname, editMetadataReq)
	r.Nil(err)
	volinfo, err = client.Volumes(volname)
	r.Nil(err)
	err = validateVolumeEdit(volinfo[0], editMetadataReq, resp)
	r.Nil(err)
	editMetadataReq = api.VolEditReq{
		Metadata: map[string]string{
			"owner": "gd2functests",
			"year":  "",
		},
		DeleteMetadata: true,
	}
	resp, err = client.EditVolume(volname, editMetadataReq)
	r.Nil(err)
	volinfo, err = client.Volumes(volname)
	r.Nil(err)
	err = validateVolumeEdit(volinfo[0], editMetadataReq, resp)
	r.Nil(err)
}
