package e2e

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	shdapi "github.com/gluster/glusterd2/plugins/glustershd/api"

	"github.com/stretchr/testify/require"
)

func checkForPendingHeals(healInfo *shdapi.BrickHealInfo) error {
	if *healInfo.EntriesInHealPending != 0 && *healInfo.EntriesInHealPending != -1 {
		return errors.New("expecting no pending heals, found pending heals")
	}
	return nil
}

func testSelfHeal(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string

	//glustershd pid file path
	pidpath := path.Join(tc.gds[0].Rundir, "glustershd.pid")

	for i := 1; i <= 2; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	reqVol := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	vol1, err := client.VolumeCreate(reqVol)
	r.Nil(err)

	r.Nil(client.VolumeStart(vol1.Name, false), "volume start failed")

	checkFuseAvailable(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)
	err = mountVolume(host, volname, mntPath)
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	defer syscall.Unmount(mntPath, syscall.MNT_FORCE)

	f, err := os.Create(mntPath + "/file1.txt")
	r.Nil(err, fmt.Sprintf("file creation failed: %s", err))
	f.Close()

	getBricksStatus, err := client.BricksStatus(volname)
	r.Nil(err, fmt.Sprintf("brick status operation failed: %s", err))
	count := 0
	for brick := range getBricksStatus {
		if getBricksStatus[brick].Info.PeerID.String() == tc.gds[0].PeerID() {
			count++
		}
	}

	r.Equal(count, 2)

	for brick := range getBricksStatus {
		if getBricksStatus[brick].Info.PeerID.String() == tc.gds[0].PeerID() {
			process, err := os.FindProcess(getBricksStatus[brick].Pid)
			r.Nil(err, fmt.Sprintf("failed to find bricks pid: %s", err))
			err = process.Signal(syscall.Signal(15))
			r.Nil(err, fmt.Sprintf("failed to kill bricks: %s", err))
			break
		}
	}

	f1, err := os.OpenFile(mntPath+"/file1.txt", os.O_WRONLY, 0222)
	r.Nil(err, fmt.Sprintf("failed to open file: %s", err))
	_, err = f1.WriteString("hello")
	r.Nil(err, fmt.Sprintf("failed to write to file: %s", err))
	f1.Sync()
	defer f1.Close()

	_, err = client.SelfHealInfo(vol1.Name)
	r.Nil(err)
	_, err = client.SelfHealInfo(vol1.Name, "split-brain-info")
	r.Nil(err)
	healInfo, err := client.SelfHealInfo(vol1.Name, "info-summary")
	r.Nil(err)

	count = 0
	for node := range healInfo {
		if healInfo[node].Status == "Connected" {
			count++
		}
	}

	r.Equal(count, 1)

	for node := range healInfo {
		if healInfo[node].Status == "Connected" {
			r.NotNil(checkForPendingHeals(&healInfo[node]))
		}
	}

	r.Nil(client.VolumeStop(vol1.Name), "Volume stop failed")
	r.Nil(client.VolumeStart(vol1.Name, false), "volume start failed")

	var optionReq api.VolOptionReq

	optionReq.Options = map[string]string{"cluster/replicate.self-heal-daemon": "on"}
	optionReq.AllowAdvanced = true

	r.Nil(client.VolumeSet(vol1.Name, optionReq))
	r.True(isProcessRunning(pidpath), "glustershd is not running")

	r.Nil(client.SelfHeal(vol1.Name, "index"))

	healInfo, err = client.SelfHealInfo(vol1.Name, "info-summary")
	r.Nil(err)

	count = 0
	for node := range healInfo {
		if healInfo[node].Status == "Connected" {
			count++
		}
	}

	r.Equal(count, 2)

	for node := range healInfo {
		if healInfo[node].Status == "Connected" {
			r.Nil(checkForPendingHeals(&healInfo[node]))
		}
	}

	r.Nil(client.SelfHeal(vol1.Name, "full"))

	// Stop Volume
	r.Nil(client.VolumeStop(vol1.Name), "Volume stop failed")

	optionReq.Options = map[string]string{"cluster/replicate.self-heal-daemon": "off"}
	optionReq.AllowAdvanced = true

	r.Nil(client.VolumeSet(vol1.Name, optionReq))
	r.False(isProcessRunning(pidpath), "glustershd is still running")

	// delete volume
	r.Nil(client.VolumeDelete(vol1.Name))
}

func testGranularEntryHeal(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string
	pidpath := path.Join(tc.gds[0].Rundir, "glustershd.pid")

	for i := 1; i <= 2; i++ {
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
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	_, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeStart(volname, false), "volume start failed")

	healInfo, err := client.SelfHealInfo(volname, "info-summary")
	r.Nil(err)
	for node := range healInfo {
		if healInfo[node].Status == "Connected" {
			r.Nil(checkForPendingHeals(&healInfo[node]))
		}
	}

	var optionReq api.VolOptionReq
	optionReq.Options = map[string]string{"cluster/replicate.granular-entry-heal": "enable"}
	optionReq.AllowAdvanced = true
	r.Nil(client.VolumeSet(volname, optionReq))

	optionReq.Options = map[string]string{"cluster/replicate.self-heal-daemon": "off"}
	optionReq.AllowAdvanced = true
	r.Nil(client.VolumeSet(volname, optionReq))
	r.False(isProcessRunning(pidpath), "glustershd is still running")

	checkFuseAvailable(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)
	err = mountVolume(host, volname, mntPath)
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	defer syscall.Unmount(mntPath, syscall.MNT_FORCE)

	getBricksStatus, err := client.BricksStatus(volname)
	r.Nil(err, fmt.Sprintf("brick status operation failed: %s", err))
	for brick := range getBricksStatus {
		if getBricksStatus[brick].Info.PeerID.String() == tc.gds[0].PeerID() {
			process, err := os.FindProcess(getBricksStatus[brick].Pid)
			r.Nil(err, fmt.Sprintf("failed to find bricks pid: %s", err))
			err = process.Signal(syscall.Signal(15))
			r.Nil(err, fmt.Sprintf("failed to kill bricks: %s", err))
			break
		}
	}

	f, err := os.Create(mntPath + "/file1.txt")
	r.Nil(err, fmt.Sprintf("file creation failed: %s", err))
	f.Close()

	healInfo, err = client.SelfHealInfo(volname, "info-summary")
	r.Nil(err)
	for node := range healInfo {
		if healInfo[node].Status == "Connected" {
			r.NotNil(checkForPendingHeals(&healInfo[node]))
		}
	}

	optionReq.Options = map[string]string{"cluster/replicate.granular-entry-heal": "disable"}
	optionReq.AllowAdvanced = true
	r.Nil(client.VolumeSet(volname, optionReq))

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	r.Nil(client.VolumeStart(volname, false), "volume start failed")

	optionReq.Options = map[string]string{"cluster/replicate.granular-entry-heal": "enable"}
	optionReq.AllowAdvanced = true
	r.NotNil(client.VolumeSet(volname, optionReq))

	err = syscall.Unmount(mntPath, 0)
	r.Nil(err)

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	// delete volume
	r.Nil(client.VolumeDelete(volname))
}

func testSplitBrainOperation(t *testing.T, tc *testCluster) {
	r := require.New(t)

	var brickPaths []string
	for i := 1; i <= 2; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	volname := formatVolName(t.Name())

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	_, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeStart(volname, false), "volume start failed")

	var optionReq api.VolOptionReq
	pidpath := path.Join(tc.gds[0].Rundir, "glustershd.pid")
	optionReq.Options = map[string]string{"cluster/replicate.self-heal-daemon": "off"}
	optionReq.AllowAdvanced = true
	r.Nil(client.VolumeSet(volname, optionReq))
	r.False(isProcessRunning(pidpath), "glustershd is still running")

	if _, err := os.Lstat("/dev/fuse"); os.IsNotExist(err) {
		t.Skip("skipping mount /dev/fuse unavailable")
	}

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)
	err = mountVolume(host, volname, mntPath)
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))
	defer syscall.Unmount(mntPath, syscall.MNT_FORCE|syscall.MNT_DETACH)

	f, err := os.Create(mntPath + "/file1.txt")
	r.Nil(err, fmt.Sprintf("file creation failed: %s", err))
	f.Close()

	var prevKilledBrick string
	getBricksStatus, err := client.BricksStatus(volname)
	r.Nil(err, fmt.Sprintf("brick status operation failed: %s", err))

	for brick := range getBricksStatus {
		if getBricksStatus[brick].Info.PeerID.String() == tc.gds[0].PeerID() {
			prevKilledBrick = getBricksStatus[brick].Info.Path
			process, err := os.FindProcess(getBricksStatus[brick].Pid)
			r.Nil(err, fmt.Sprintf("failed to find bricks pid: %s", err))
			err = process.Signal(syscall.Signal(15))
			r.Nil(err, fmt.Sprintf("failed to kill bricks: %s", err))
			break
		}
	}

	f1, err := os.OpenFile(mntPath+"/file1.txt", os.O_RDWR, 0777)
	r.Nil(err, fmt.Sprintf("failed to open file: %s", err))
	_, err = f1.WriteString("hello")
	r.Nil(err, fmt.Sprintf("failed to write to file: %s", err))
	f1.Sync()
	defer f1.Close()

	err = syscall.Unmount(mntPath, syscall.MNT_FORCE|syscall.MNT_DETACH)
	r.Nil(err)

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	// Start Volume
	r.Nil(client.VolumeStart(volname, false), "Volume start failed")

	err = mountVolume(host, volname, mntPath)
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))
	defer syscall.Unmount(mntPath, syscall.MNT_FORCE|syscall.MNT_DETACH)

	getBricksStatus, err = client.BricksStatus(volname)
	r.Nil(err, fmt.Sprintf("brick status operation failed: %s", err))

	for brick := range getBricksStatus {
		if getBricksStatus[brick].Info.PeerID.String() == tc.gds[0].PeerID() && getBricksStatus[brick].Info.Path != prevKilledBrick {
			process, err := os.FindProcess(getBricksStatus[brick].Pid)
			r.Nil(err, fmt.Sprintf("failed to find bricks pid: %s", err))
			err = process.Signal(syscall.Signal(15))
			r.Nil(err, fmt.Sprintf("failed to kill bricks: %s", err))
			break
		}
	}

	f2, err := os.OpenFile(mntPath+"/file1.txt", os.O_RDWR, 0777)
	r.Nil(err, fmt.Sprintf("failed to open file: %s", err))
	_, err = f2.WriteString("hey")
	r.Nil(err, fmt.Sprintf("failed to write to file: %s", err))
	f2.Sync()
	defer f2.Close()

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	// Start Volume
	r.Nil(client.VolumeStart(volname, false), "Volume start failed")

	healInfo, err := client.SelfHealInfo(volname, "info-summary")
	r.Nil(err)

	r.Equal(*healInfo[0].EntriesInSplitBrain, int64(1))

	var req shdapi.SplitBrainReq

	req.FileName = ""
	err = client.SelfHealSplitBrain(volname, "latest-mtime", req)
	r.NotNil(err)

	req.FileName = "file1.txt"
	err = client.SelfHealSplitBrain(volname, "latest-mtime", req)
	r.NotNil(err)

	req.FileName = "/file1.txt"
	err = client.SelfHealSplitBrain(volname, "latest-mtime", req)
	r.Nil(err)

	healInfo, err = client.SelfHealInfo(volname, "info-summary")
	r.Nil(err)

	r.Equal(*healInfo[0].EntriesInSplitBrain, int64(0))

	err = syscall.Unmount(mntPath, syscall.MNT_FORCE|syscall.MNT_DETACH)
	r.Nil(err)

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	// delete volume
	r.Nil(client.VolumeDelete(volname))
}
