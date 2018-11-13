package e2e

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/gluster/glusterd2/e2e/lvmtest"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

var (
	snapname     = "snaptest"
	clonename    = "clonevolume"
	snapTestName string
)

// TestSnapshot creates a volume and snapshots, runs further tests on it and
// finally deletes the volume
func TestSnapshot(t *testing.T) {
	var err error
	var brickPaths []string
	brickCount := 4
	snapTestName = t.Name()
	r := require.New(t)

	prefix := testTempDir(t, "bricks")
	lvmtest.Cleanup(baseLocalStateDir, prefix, brickCount)
	defer func() {
		lvmtest.Cleanup(baseLocalStateDir, prefix, brickCount)
	}()
	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err = initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	brickPaths, err = lvmtest.CreateLvmBricks(prefix, brickCount)
	r.Nil(err)
	defer func() {
		lvmtest.CleanupLvmBricks(prefix, brickCount)
	}()

	// Create the volume
	if err := volumeCreateOnLvm(snapTestName, brickPaths, client, tc); err != nil {
		r.Nil(err, "Failed to create the volume")
	}

	defer func() {
		client.VolumeDelete(snapTestName)
		r.Nil(err)
	}()

	if err := client.VolumeStart(snapTestName, true); err != nil {
		r.Nil(err, "Failed to start the volume")
	}

	defer func() {
		err := client.VolumeStop(snapTestName)
		r.Nil(err)
	}()

	t.Run("Create", testSnapshotCreate)
	t.Run("Activate", testSnapshotActivate)
	t.Run("List", testSnapshotList)
	t.Run("Mount", tc.wrap(testSnapshotMount))
	t.Run("StatusAndForceActivate", testSnapshotStatusForceActivate)
	t.Run("Info", testSnapshotInfo)
	t.Run("Clone", testSnapshotClone)
	t.Run("Restore", testSnapshotRestore)
	t.Run("MountRestoredVolume", tc.wrap(testRestoredVolumeMount))
	t.Run("Validate", testSnapshotValidation)
	t.Run("Deactivate", testSnapshotDeactivate)
	t.Run("Delete", testSnapshotDelete)
	t.Run("DeleteClone", testCloneDelete)
	t.Run("SmartVolume", tc.wrap(testSnapshotOnSmartVol))

	/*
		TODO:
		test snapshot on a volume that is expanded or shrunk
	*/

}

func volumeCreateOnLvm(volName string, brickPaths []string, client *restclient.Client, tc *testCluster) error {

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
		Name: volName,
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
	return err
}

func testSnapshotCreate(t *testing.T) {
	r := require.New(t)
	snapshotCreateReq := api.SnapCreateReq{
		VolName:  snapTestName,
		SnapName: snapname,
	}
	_, err := client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err, "snapshot create failed")

	snapshotCreateReq = api.SnapCreateReq{
		VolName:     snapTestName,
		SnapName:    snapname,
		TimeStamp:   true,
		Description: "Snapshot for testing, timestamp and force flags are enabled",
		Force:       true,
	}
	_, err = client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err, "snapshot create failed")

}

func testSnapshotClone(t *testing.T) {
	r := require.New(t)
	snapshotCloneReq := api.SnapCloneReq{
		CloneName: clonename,
	}
	_, err := client.SnapshotClone(snapname, snapshotCloneReq)
	r.Nil(err, "snapshot clone failed")

	err = client.VolumeStart(clonename, true)
	r.Nil(err, "Failed to start cloned volume")

	volumes, err := client.Volumes("")
	r.Nil(err)
	r.Len(volumes, 2)
}

func testSnapshotList(t *testing.T) {
	r := require.New(t)

	snaps, err := client.SnapshotList("")
	r.Nil(err)
	r.Len(snaps, 1)
	r.Len(snaps[0].SnapList, 2)

	snaps, err = client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(snaps[0].SnapList, 2)

}

func testSnapshotInfo(t *testing.T) {
	r := require.New(t)

	//gets the info of the snapname without timestamp
	_, err := client.SnapshotInfo(snapname)
	r.Nil(err)
}

func testSnapshotActivate(t *testing.T) {
	var snapshotActivateReq api.SnapActivateReq
	r := require.New(t)

	vols, err := client.SnapshotList(snapTestName)
	r.Nil(err)

	for _, snaps := range vols {
		for _, snap := range snaps.SnapList {
			err = client.SnapshotActivate(snapshotActivateReq, snap.VolInfo.Name)
			r.Nil(err)

			snapshotActivateReq.Force = true
		}
	}

}

func testSnapshotDelete(t *testing.T) {
	r := require.New(t)

	vols, err := client.SnapshotList("")
	r.Nil(err)
	r.Len(vols[0].SnapList, 1)

	for _, snaps := range vols {
		for _, snap := range snaps.SnapList {
			err = client.SnapshotDelete(snap.VolInfo.Name)
			r.Nil(err)
		}
	}

	vols, err = client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(vols, 0)
}

func testSnapshotDeactivate(t *testing.T) {
	r := require.New(t)
	vols, err := client.SnapshotList(snapTestName)
	r.Nil(err)

	for _, snaps := range vols {
		for _, snap := range snaps.SnapList {
			err = client.SnapshotDeactivate(snap.VolInfo.Name)
			r.Nil(err)
		}
	}

}

func testSnapshotStatusForceActivate(t *testing.T) {
	var snapshotActivateReq api.SnapActivateReq
	var result api.SnapStatusResp
	r := require.New(t)

	vols, err := client.SnapshotList(snapTestName)
	r.Nil(err)

	snapName := vols[0].SnapList[0].VolInfo.Name
	result, err = client.SnapshotStatus(snapName)
	r.Nil(err)

	err = daemon.Kill(result.BrickStatus[0].Brick.Pid, true)
	r.Nil(err)
	err = client.SnapshotActivate(snapshotActivateReq, snapName)
	if err == nil {
		msg := "snapshot activate should have failed"
		r.Nil(errors.New(msg), msg)
	}
	snapshotActivateReq.Force = true
	err = client.SnapshotActivate(snapshotActivateReq, snapName)
	r.Nil(err)

	retries := 4
	waitTime := 6000
	err = errors.New("snapshot failed to activate forcefully")
	for i := 0; i < retries; i++ {
		// opposite of exponential backoff
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
		result, err = client.SnapshotStatus(snapName)
		r.Nil(err)

		online := result.BrickStatus[0].Brick.Online
		if online {
			err = nil
			break
		}

		waitTime = waitTime / 2
	}
	r.Nil(err)
}

func testSnapshotRestore(t *testing.T) {
	var result api.SnapStatusResp
	r := require.New(t)

	vols, err := client.SnapshotList(snapTestName)
	r.Nil(err)

	snapName := vols[0].SnapList[0].VolInfo.Name
	result, err = client.SnapshotStatus(snapName)
	r.Nil(err)

	err = daemon.Kill(result.BrickStatus[0].Brick.Pid, true)
	r.Nil(err)
	err = client.VolumeStop(snapTestName)
	r.Nil(err)

	_, err = client.SnapshotRestore(snapName)
	r.Nil(err)

	snaps, err := client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(snaps[0].SnapList, 1)

	err = client.VolumeStart(snapTestName, true)
	r.Nil(err)

}

func testRestoredVolumeMount(t *testing.T, tc *testCluster) {
	r := require.New(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)

	err := mountVolume(host, snapTestName, mntPath)
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	defer syscall.Unmount(mntPath, syscall.MNT_FORCE)

	err = testMount(mntPath)
	r.Nil(err)

	err = syscall.Unmount(mntPath, 0)
	r.Nil(err, fmt.Sprintf("unmount failed: %s", err))
}

func testSnapshotMount(t *testing.T, tc *testCluster) {
	r := require.New(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)
	volID := fmt.Sprintf("/snaps/%s", snapname)

	err := mountVolume(host, volID, mntPath)
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	defer syscall.Unmount(mntPath, syscall.MNT_FORCE)

	newDir := mntPath + "/Dir"
	err = syscall.Mkdir(newDir, 0755)
	if err == nil {
		r.Nil(errors.New("snapshot volume is Read Only File System"))
	}

	err = syscall.Unmount(mntPath, 0)
	r.Nil(err, fmt.Sprintf("unmount failed: %s", err))
}

func testSnapshotValidation(t *testing.T) {
	r := require.New(t)

	snapshotCreateReq := api.SnapCreateReq{
		VolName:   clonename,
		SnapName:  snapname,
		TimeStamp: true,
	}
	_, err := client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err, "snapshot create failed")

	r.Nil(client.VolumeStop(clonename), "volume stop failed")

	err = client.VolumeDelete(clonename)
	r.NotNil(err, "Volume delete succeeded when snapshot is existing for the volume")
}

func testCloneDelete(t *testing.T) {
	r := require.New(t)

	r.Nil(client.VolumeDelete(clonename))

	volumes, err := client.Volumes("")
	r.Nil(err)
	r.Len(volumes, 1)

	//TODO Test lv device removed or not
}

func testSnapshotOnSmartVol(t *testing.T, tc *testCluster) {
	r := require.New(t)

	devicesDir := testTempDir(t, "devices")
	r.Nil(prepareLoopDevice(devicesDir+"/gluster_dev1.img", "1", "500M"))
	r.Nil(prepareLoopDevice(devicesDir+"/gluster_dev2.img", "2", "500M"))

	_, err := client.DeviceAdd(tc.gds[0].PeerID(), "/dev/gluster_loop1")
	r.Nil(err)

	_, err = client.DeviceAdd(tc.gds[1].PeerID(), "/dev/gluster_loop2")
	r.Nil(err)

	smartvolname := formatVolName(t.Name())
	// create Replica 2 Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         200,
		ReplicaCount: 2,
	}
	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeStart(smartvolname, true))

	snapshotCreateReq := api.SnapCreateReq{
		VolName:  smartvolname,
		SnapName: snapname,
	}
	//Create snapshot with name snapname (Previous snaps with same name is already deleted)
	//This also tests snapshot create with same name after a deletion
	_, err = client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err)

	snapshotActivateReq := api.SnapActivateReq{
		Force: true,
	}
	r.Nil(client.SnapshotActivate(snapshotActivateReq, snapname))

	//Creating a clone from the snapshot
	snapshotCloneReq := api.SnapCloneReq{
		CloneName: clonename,
	}
	_, err = client.SnapshotClone(snapname, snapshotCloneReq)
	r.Nil(err, "snapshot clone failed")

	r.Nil(client.VolumeStop(smartvolname))

	//Restoring the snapshot to the parent volume
	//During this process, parent volume thinLV should delete
	_, err = client.SnapshotRestore(snapname)
	r.Nil(err)

	r.Nil(client.VolumeStart(smartvolname, true))

	r.Nil(client.VolumeDelete(clonename))

	r.Nil(client.VolumeStop(smartvolname))

	r.Nil(client.VolumeDelete(smartvolname))

	//At this point all snapshot and volumes are deleted.
	//So the lvcount should be zero
	checkZeroLvs(r)

	r.Nil(loopDevicesCleanup(t))

}
