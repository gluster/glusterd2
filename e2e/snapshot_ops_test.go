package e2e

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
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

	prefix := fmt.Sprintf("%s/%s/bricks/", baseLocalStateDir, snapTestName)
	lvmtest.Cleanup(baseLocalStateDir, prefix, brickCount)
	defer func() {
		lvmtest.Cleanup(baseLocalStateDir, prefix, brickCount)
	}()
	tc, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client = initRestclient(tc.gds[0])
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
		client.VolumeStop(snapTestName)
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
	t.Run("Deactivate", testSnapshotDeactivate)
	t.Run("Delete", testSnapshotDelete)

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

	r.Nil(client.VolumeStop(clonename), "volume stop failed")

	r.Nil(client.VolumeDelete(clonename))

	volumes, err = client.Volumes("")
	r.Nil(err)
	r.Len(volumes, 1)
}

func testSnapshotList(t *testing.T) {
	r := require.New(t)

	snaps, err := client.SnapshotList("")
	r.Nil(err)
	r.Len(snaps[0].SnapName, 2)

	snaps, err = client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(snaps[0].SnapName, 2)

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
		for _, snapName := range snaps.SnapName {
			err = client.SnapshotActivate(snapshotActivateReq, snapName)
			r.Nil(err)

			snapshotActivateReq.Force = true
		}
	}

}

func testSnapshotDelete(t *testing.T) {
	r := require.New(t)

	vols, err := client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(vols[0].SnapName, 1)

	for _, snaps := range vols {
		for _, snapName := range snaps.SnapName {
			err = client.SnapshotDelete(snapName)
			r.Nil(err)
		}
	}

	vols, err = client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(vols[0].SnapName, 0)
}

func testSnapshotDeactivate(t *testing.T) {
	r := require.New(t)
	vols, err := client.SnapshotList(snapTestName)
	r.Nil(err)

	for _, snaps := range vols {
		for _, snapName := range snaps.SnapName {
			err = client.SnapshotDeactivate(snapName)
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

	snapName := vols[0].SnapName[0]
	result, err = client.SnapshotStatus(snapName)
	r.Nil(err)

	err = daemon.Kill(result.BrickStatus[0].Brick.Pid, true)
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

	snapName := vols[0].SnapName[0]
	result, err = client.SnapshotStatus(snapName)
	r.Nil(err)

	err = daemon.Kill(result.BrickStatus[0].Brick.Pid, true)
	err = client.VolumeStop(snapTestName)
	r.Nil(err)

	_, err = client.SnapshotRestore(snapName)
	r.Nil(err)

	snaps, err := client.SnapshotList(snapTestName)
	r.Nil(err)
	r.Len(snaps[0].SnapName, 1)

	err = client.VolumeStart(snapTestName, true)
	r.Nil(err)

}

func testRestoredVolumeMount(t *testing.T, tc *testCluster) {
	r := require.New(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)
	mntCmd := exec.Command("mount", "-t", "glusterfs", host+":"+snapTestName, mntPath)
	umntCmd := exec.Command("umount", mntPath)

	err := mntCmd.Run()
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	err = umntCmd.Run()
	r.Nil(err, fmt.Sprintf("unmount failed: %s", err))
}

func testSnapshotMount(t *testing.T, tc *testCluster) {
	r := require.New(t)

	mntPath := testTempDir(t, "mnt")
	defer os.RemoveAll(mntPath)

	host, _, _ := net.SplitHostPort(tc.gds[0].ClientAddress)

	volID := fmt.Sprintf("%s:/snaps/%s", host, snapname)
	mntCmd := exec.Command("mount", "-t", "glusterfs", volID, mntPath)
	umntCmd := exec.Command("umount", mntPath)

	err := mntCmd.Run()
	r.Nil(err, fmt.Sprintf("mount failed: %s", err))

	newDir := mntPath + "/Dir"
	err = syscall.Mkdir(newDir, 0755)
	if err == nil {
		r.Nil(errors.New("snapshot volume is Read Only File System"))
	}

	err = umntCmd.Run()
	r.Nil(err, fmt.Sprintf("unmount failed: %s", err))
}
