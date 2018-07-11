package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/stretchr/testify/require"
)

// TestRestart tests that data persists after a GD2 restart
func TestRestart(t *testing.T) {
	r := require.New(t)

	gd, err := spawnGlusterd("./config/1.toml", true)
	r.Nil(err)
	r.True(gd.IsRunning())

	dir, err := ioutil.TempDir(baseLocalStateDir, t.Name())
	r.Nil(err)
	defer os.RemoveAll(dir)

	client := initRestclient(gd)

	createReq := api.VolCreateReq{
		Name: formatVolName(t.Name()),
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: gd.PeerID(), Path: dir},
				},
			},
		},
		Force: true,
	}
	_, errVolCreate := client.VolumeCreate(createReq)
	r.Nil(errVolCreate)

	r.Len(getVols(gd, r), 1)

	r.Nil(gd.Stop())

	gd, err = spawnGlusterd("./config/1.toml", false)
	r.Nil(err)
	r.True(gd.IsRunning())

	r.Len(getVols(gd, r), 1)

	r.Nil(gd.Stop())
}

func getVols(gd *gdProcess, r *require.Assertions) api.VolumeListResp {
	client := initRestclient(gd)
	volname := ""
	vols, err := client.Volumes(volname)
	r.Nil(err)
	return vols
}
