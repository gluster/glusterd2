package e2e

import (
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

// TestBitrot creates a volume runs further tests on it
func TestBitrot(t *testing.T) {
	var err error

	r := require.New(t)

	tc, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client = initRestclient(tc.gds[0])

	// test Bitrot on dist-rep volume
	t.Run("Replica-volume", tc.wrap(testBitrotOnReplicaVolume))
	// test Bitrot on pure distribute volume
	t.Run("Dist-volume", tc.wrap(testBitrotOnDistVolume))

}

func testBitrotOnReplicaVolume(t *testing.T, tc *testCluster) {
	r := require.New(t)
	volumeName := formatVolName(t.Name())
	var brickPaths []string

	for i := 1; i <= 4; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
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
		Force: true,
	}

	_, _, err := client.VolumeCreate(createReq)
	r.Nil(err)

	testbitrot(t)

	_, err = client.VolumeDelete(volumeName)
	r.Nil(err)
}

func testBitrotOnDistVolume(t *testing.T, tc *testCluster) {
	r := require.New(t)
	volumeName := formatVolName(t.Name())
	var brickPaths []string

	for i := 1; i <= 4; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	createReq := api.VolCreateReq{
		Name: volumeName,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Force: true,
	}

	_, _, err := client.VolumeCreate(createReq)
	r.Nil(err)

	_, _, err = client.Volumes(volumeName)
	r.Nil(err)
	testbitrot(t)

	_, err = client.VolumeDelete(volumeName)
	r.Nil(err)

}

func testbitrot(t *testing.T) {
	volumeName := formatVolName(t.Name())
	r := require.New(t)

	//check bitrot status, before starting volume
	_, _, err1 := client.BitrotScrubStatus(volumeName)
	r.Contains(err1.Error(), "volume not started")

	//start volume
	_, err := client.VolumeStart(volumeName, true)
	r.Nil(err)

	//check bitrot status on started volume
	_, _, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "bitrot is not enabled")

	//enable bitrot on volume
	_, err = client.BitrotEnable(volumeName)
	r.Nil(err)

	//check bitrot status
	scrubStatus, _, err := client.BitrotScrubStatus(volumeName)
	r.Nil(err)
	r.Equal(scrubStatus.State, "Active (Idle)")

	//disable bitrot on volume
	_, err = client.BitrotDisable(volumeName)
	r.Nil(err)

	//check bitrot status
	_, _, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "bitrot is not enabled")

	//stop volume
	_, err = client.VolumeStop(volumeName)
	r.Nil(err)

	//check bitrot status
	_, _, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "volume not started")

	//enable bitrot on volume
	_, err = client.BitrotEnable(volumeName)
	r.Contains(err.Error(), "volume not started")

	//start volume
	_, err = client.VolumeStart(volumeName, true)
	r.Nil(err)

	//check bitrot status
	scrubStatus, _, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "bitrot is not enabled")

	//disable bitrot on volume
	_, err = client.BitrotDisable(volumeName)
	r.Contains(err.Error(), "bitrot is already disabled")

	//check bitrot status
	_, _, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "bitrot is not enabled")

	//stop volume
	_, err = client.VolumeStop(volumeName)
	r.Nil(err)

}
