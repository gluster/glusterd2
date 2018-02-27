package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

// TestVolume creates a volume and starts it, runs further quota enable on it
// and finally deletes the volume
func testQuotaEnable(t *testing.T) {
	var err error
	r := require.New(t)

	gds, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", t.Name())
	r.Nil(err)
	defer os.RemoveAll(brickDir)

	var brickPaths [4]string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths[i-1] = brickPath
	}

	client := initRestclient(gds[0].ClientAddress)
	volname1 := "testvol1"
	reqVol := api.VolCreateReq{
		Name: volname1,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{NodeID: gds[0].PeerID(), Path: brickPaths[0]},
					{NodeID: gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	vol1, err := client.VolumeCreate(reqVol)
	r.Nil(err)

	r.Nil(client.VolumeStart(vol1.Name), "volume start failed")

	err = client.QuotaEnable(volname)
	r.Nil(err)

	// Stop Volume
	r.Nil(client.VolumeStop(vol1.Name), "Volume stop failed")
	// delete volume
	err = client.VolumeDelete(vol1.Name)
	r.Nil(err)

}
