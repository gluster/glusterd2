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

	brickDir, err := ioutil.TempDir(baseLocalStateDir, t.Name())
	r.Nil(err)
	defer os.RemoveAll(brickDir)

	var brickPaths [4]string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths[i-1] = brickPath
	}

	client := initRestclient(gds[0].ClientAddress)
	volname := formatVolName(t.Name())
	reqVol := api.VolCreateReq{
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
		},
		Force: true,
	}
	_, err = client.VolumeCreate(reqVol)
	r.Nil(err)

	r.Nil(client.VolumeStart(volname, false), "volume start failed")

	err = client.QuotaEnable(volname)
	r.Nil(err)

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	// delete volume
	err = client.VolumeDelete(volname)
	r.Nil(err)

}
