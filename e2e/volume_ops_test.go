package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestVolumeCreateDelete(t *testing.T) {
	r := require.New(t)

	gds, err := setupCluster("./config/1.yaml", "./config/2.yaml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", "TestVolumeCreateDelete")
	defer os.RemoveAll(brickDir)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	client := initRestclient(gds[0].ClientAddress)

	// create 2x2 dist-rep volume
	volname := "testvol"
	createReq := api.VolCreateReq{
		Name:    volname,
		Replica: 2,
		Bricks: []string{
			gds[0].PeerAddress + ":" + brickPaths[0],
			gds[1].PeerAddress + ":" + brickPaths[1],
			gds[0].PeerAddress + ":" + brickPaths[2],
			gds[1].PeerAddress + ":" + brickPaths[3]},
		Force: true,
	}
	_, errVolCreate := client.VolumeCreate(createReq)
	r.Nil(errVolCreate)

	// delete volume
	errVolDel := client.VolumeDelete(volname)
	r.Nil(errVolDel)
}
