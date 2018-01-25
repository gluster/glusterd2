package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestGlusterShd(t *testing.T) {

	r := require.New(t)

	gds, err := setupCluster("./config/1.yaml", "./config/2.yaml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", "TestShdEnable")
	r.Nil(err)
	defer os.RemoveAll(brickDir)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
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

	err = client.GlusterShdEnable(vol1.Name)
	r.Nil(err)

	err = client.GlusterShdDisable(vol1.Name)
	r.Nil(err)

	// delete volume
	err = client.VolumeDelete(vol1.Name)
	r.Nil(err)

}
