package e2e

import (
	"fmt"
	"testing"

	"github.com/gluster/glusterd2/e2e/lvmtest"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

var (
	snapname     = "snaptest"
	snapTestName string
)

// TestSnapshot creates a volume and snapshots, runs further tests on it and
// finally deletes the volume
func TestSnapshot(t *testing.T) {
	var err error
	var brickPaths []string
	snapTestName = t.Name()
	r := require.New(t)

	gds, err = setupCluster("./config/1.toml", "./config/2.toml")
	if err != nil {
		r.Nil(err)
		teardownCluster(gds)
		return
	}
	defer teardownCluster(gds)

	client = initRestclient(gds[0].ClientAddress)
	prefix := fmt.Sprintf("%s/%s/bricks/", baseWorkdir, snapTestName)
	brickPaths, err = lvmtest.CreateLvmBricks(prefix, 4)
	if err != nil {
		r.Nil(err)
		lvmtest.CleanupLvmBricks(4)
		return
	}
	defer func() {
		err := lvmtest.CleanupLvmBricks(4)
		r.Nil(err)
	}()

	// Create the volume
	if err := volumeCreateOnLvm(snapTestName, brickPaths, client); err != nil {
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
	t.Run("Deactivate", testSnapshotDeactivate)
	t.Run("List", testSnapshotList)
	t.Run("Info", testSnapshotInfo)
	t.Run("Delete", testSnapshotDelete)

	/*
		TODO:
		test snapshot on a volume that is expanded or shrinked
	*/

}

func volumeCreateOnLvm(volName string, brickPaths []string, client *restclient.Client) error {

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
		Name: volName,
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
	r.Len(vols[0].SnapName, 2)

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
