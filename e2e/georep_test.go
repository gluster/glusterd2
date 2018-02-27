package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
	"github.com/stretchr/testify/require"
)

func TestGeorepCreateDelete(t *testing.T) {
	r := require.New(t)

	gds, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", "TestGeorepCreate")
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
				Type: "distribute",
				Bricks: []api.BrickReq{
					{NodeID: gds[0].PeerID(), Path: brickPaths[0]},
					{NodeID: gds[0].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	vol1, err := client.VolumeCreate(reqVol)
	r.Nil(err)

	volname2 := "testvol2"
	reqVol = api.VolCreateReq{
		Name: volname2,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{NodeID: gds[1].PeerID(), Path: brickPaths[2]},
					{NodeID: gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Force: true,
	}
	vol2, err := client.VolumeCreate(reqVol)
	r.Nil(err)

	reqGeorep := georepapi.GeorepCreateReq{
		MasterVol: volname1,
		RemoteVol: volname2,
		RemoteHosts: []georepapi.GeorepRemoteHostReq{
			{NodeID: gds[1].PeerID(), Hostname: gds[1].PeerAddress},
		},
	}

	_, err = client.GeorepCreate(vol1.ID.String(), vol2.ID.String(), reqGeorep)
	r.Nil(err)

	// delete geo-rep session
	err = client.GeorepDelete(vol1.ID.String(), vol2.ID.String(), false)
	r.Nil(err)

	// delete volume
	err = client.VolumeDelete(volname1)
	r.Nil(err)

	// delete volume
	err = client.VolumeDelete(volname2)
	r.Nil(err)
}
