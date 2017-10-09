package e2e

import (
	"io/ioutil"
	"os"
	"os/exec"
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
	r.Nil(err)
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
			gds[0].PeerID() + ":" + brickPaths[0],
			gds[1].PeerID() + ":" + brickPaths[1],
			gds[0].PeerID() + ":" + brickPaths[2],
			gds[1].PeerID() + ":" + brickPaths[3]},
		Force: true,
	}
	_, errVolCreate := client.VolumeCreate(createReq)
	r.Nil(errVolCreate)

	// delete volume
	errVolDel := client.VolumeDelete(volname)
	r.Nil(errVolDel)
}

func TestVolumeOptions(t *testing.T) {

	// skip this test if glusterfs server packages and xlators are not
	// installed
	_, err := exec.Command("sh", "-c", "which glusterfsd").Output()
	if err != nil {
		t.SkipNow()
	}

	r := require.New(t)

	gds, err := setupCluster("./config/1.yaml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", t.Name())
	defer os.RemoveAll(brickDir)

	brickPath, err := ioutil.TempDir(brickDir, "brick")
	r.Nil(err)

	client := initRestclient(gds[0].ClientAddress)

	volname := "testvol"
	createReq := api.VolCreateReq{
		Name:   volname,
		Bricks: []string{gds[0].PeerID() + ":" + brickPath},
		Force:  true,
	}

	// valid option test cases
	validOpKeys := []string{"gfproxy.afr.eager-lock", "afr.eager-lock"}
	for _, validKey := range validOpKeys {
		createReq.Options = map[string]string{validKey: "on"}

		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		err = client.VolumeDelete(volname)
		r.Nil(err)
	}

	// invalid option test cases
	invalidOpKeys := []string{"..eager-lock", "a.b.afr.eager-lock", "afr.non-existent", "eager-lock"}
	for _, invalidKey := range invalidOpKeys {
		createReq.Options = map[string]string{invalidKey: "on"}

		_, err = client.VolumeCreate(createReq)
		r.NotNil(err)
	}
}
