package e2e

import (
	"io/ioutil"
	"testing"

	"github.com/gluster/glusterd2/e2e/lvmtest"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

var (
	snapname = "snaptest"
)

// TestSnapshot creates a volume and snapshots, runs further tests on it and
// finally deletes the volume
func TestSnapshot(t *testing.T) {
	var err error
	var brickPaths []string
	r := require.New(t)

	gds, err = setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	client = initRestclient(gds[0].ClientAddress)
	brickPaths, err = lvmtest.CreateLvmBricks(4)
	if err != nil {
		r.Nil(err)
		return
	}
	defer func() {
		err := lvmtest.CleanupLvmBricks(4)
		r.Nil(err)
	}()

	// Create the volume
	if err := volumeCreateOnLvm(brickPaths, client); err != nil {
		t.Logf("Failed to create the volume")
		r.Nil(err)
		return
	}

	defer func() {
		err := volumeDelete(client)
		r.Nil(err)
	}()

	if err := volumeStart(client); err != nil {
		t.Logf("Failed to start the volume")
		r.Nil(err)
		return
	}

	defer func() {
		err := volumeStop(client)
		r.Nil(err)
	}()

	t.Run("Create", testSnapshotCreate)
	t.Run("Activate", testSnapshotActivate)
	t.Run("Deactivate", testSnapshotDeactivate)
	t.Run("List", testSnapshotList)
	t.Run("Info", testSnapshotInfo)

	/*
		TODO:
		test snapshot on a volume that is expanded or shrinked
	*/

}

func createLvmBricks(number int) ([]string, error) {
	/*
		if err := verifyLvm(); err != nil {
			return err
		}
	*/
	var brickPaths []string
	for i := 1; i <= number; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		if err != nil {
			return brickPaths, err
		}
		brickPaths = append(brickPaths, brickPath)
	}
	return brickPaths, nil
}

func volumeCreateOnLvm(brickPaths []string, client *restclient.Client) error {

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
	return err
}

func volumeDelete(client *restclient.Client) error {
	return client.VolumeDelete(volname)
}

func testSnapshotCreate(t *testing.T) {
	r := require.New(t)
	snapshotCreateReq := api.SnapCreateReq{
		VolName:  volname,
		SnapName: snapname,
	}
	_, err := client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err, "snapshot create failed")

	snapshotCreateReq = api.SnapCreateReq{
		VolName:     volname,
		SnapName:    snapname,
		TimeStamp:   true,
		Description: "Snapshot for testing, timestamp and force flags are enabled",
		Force:       true,
	}
	_, err = client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err, "snapshot create failed")

}

func volumeStart(client *restclient.Client) error {

	return client.VolumeStart(volname, true)
}

func volumeStop(client *restclient.Client) error {

	return client.VolumeStop(volname)
}

func testSnapshotList(t *testing.T) {
	var snapshotListReq api.SnapListReq
	r := require.New(t)

	snaps, err := client.SnapshotList(snapshotListReq)
	r.Nil(err)
	r.Len(snaps[0].SnapName, 2)
	snapshotListReq.Volname = volname
	snaps, err = client.SnapshotList(snapshotListReq)
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
	var snapshotListReq api.SnapListReq
	var snapshotActivateReq api.SnapActivateReq
	r := require.New(t)

	snapshotListReq.Volname = volname
	vols, err := client.SnapshotList(snapshotListReq)
	r.Nil(err)

	for _, snaps := range vols {
		for _, snapName := range snaps.SnapName {
			err = client.SnapshotActivate(snapshotActivateReq, snapName)
			r.Nil(err)

			snapshotActivateReq.Force = true
		}
	}

}

func testSnapshotDeactivate(t *testing.T) {
	var snapshotListReq api.SnapListReq
	var snapshotActivateReq api.SnapActivateReq
	r := require.New(t)

	snapshotListReq.Volname = volname
	vols, err := client.SnapshotList(snapshotListReq)
	r.Nil(err)
	for _, snaps := range vols {
		for _, snapName := range snaps.SnapName {
			err = client.SnapshotDeactivate(snapName)
			r.Nil(err)

			snapshotActivateReq.Force = true
		}
	}

}
