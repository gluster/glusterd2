package e2e

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

// TestBitrot creates a volume runs further tests on it
func TestBitrot(t *testing.T) {
	var err error

	r := require.New(t)

	gds, err = setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	client = initRestclient(gds[0].ClientAddress)

	tmpDir, err = ioutil.TempDir(baseWorkdir, t.Name())
	r.Nil(err)
	t.Logf("Using temp dir: %s", tmpDir)

	// test Bitrot on dist-rep volume
	t.Run("Replica-volume", testBitrotOnReplicaVolume)
	// test Bitrot on pure distribute volume
	t.Run("Dist-volume", testBitrotOnDistVolume)

}

func testBitrotOnReplicaVolume(t *testing.T) {
	r := require.New(t)
	volumeName := strings.Replace(t.Name(), "/", "-", 1)
	var brickPaths []string

	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		r.Nil(err)
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

	testbitrot(t)

	r.Nil(client.VolumeDelete(volumeName))
}

func testBitrotOnDistVolume(t *testing.T) {
	r := require.New(t)
	volumeName := strings.Replace(t.Name(), "/", "-", 1)
	var brickPaths []string

	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	createReq := api.VolCreateReq{
		Name: volumeName,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type: "distribute",
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

	_, err = client.Volumes(volumeName)
	r.Nil(err)
	testbitrot(t)

	r.Nil(client.VolumeDelete(volumeName))

}

func testbitrot(t *testing.T) {
	volumeName := strings.Replace(t.Name(), "/", "-", 1)
	r := require.New(t)

	//check bitrot status, before starting volume
	_, err1 := client.BitrotScrubStatus(volumeName)
	r.Contains(err1.Error(), "volume not started")

	//start volume
	err := client.VolumeStart(volumeName, true)
	r.Nil(err)

	//check bitrot status on started volume
	_, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "Bitrot is not enabled")

	//enable bitrot on volume
	err = client.BitrotEnable(volumeName)
	r.Nil(err)

	//check bitrot status
	scrubStatus, err := client.BitrotScrubStatus(volumeName)
	r.Nil(err)
	r.Equal(scrubStatus.State, "Active (Idle)")

	//disable bitrot on volume
	err = client.BitrotDisable(volumeName)
	r.Nil(err)

	//check bitrot status
	_, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "Bitrot is not enabled")

	//stop volume
	err = client.VolumeStop(volumeName)
	r.Nil(err)

	//check bitrot status
	_, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "volume not started")

	//enable bitrot on volume
	err = client.BitrotEnable(volumeName)
	r.Contains(err.Error(), "volume not started")

	//start volume
	err = client.VolumeStart(volumeName, true)
	r.Nil(err)

	//check bitrot status
	scrubStatus, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "Bitrot is not enabled")

	//disable bitrot on volume
	err = client.BitrotDisable(volumeName)
	r.Contains(err.Error(), "Bitrot is already disabled")

	//check bitrot status
	_, err = client.BitrotScrubStatus(volumeName)
	r.Contains(err.Error(), "Bitrot is not enabled")

	//stop volume
	err = client.VolumeStop(volumeName)
	r.Nil(err)

}
