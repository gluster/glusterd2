package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestSelfHeal(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster("./config/1.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	brickDir, err := ioutil.TempDir(baseLocalStateDir, t.Name())
	r.Nil(err)
	defer os.RemoveAll(brickDir)

	var brickPaths [4]string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths[i-1] = brickPath
	}

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	volname := formatVolName(t.Name())
	reqVol := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[1]},
				},
			},
		},
		Force: true,
	}
	vol1, err := client.VolumeCreate(reqVol)
	r.Nil(err)

	r.Nil(client.VolumeStart(vol1.Name, false), "volume start failed")

	_, err = client.SelfHealInfo(vol1.Name)
	r.Nil(err)
	_, err = client.SelfHealInfo(vol1.Name, "info-summary")
	r.Nil(err)
	_, err = client.SelfHealInfo(vol1.Name, "split-brain-info")
	r.Nil(err)

	var optionReq api.VolOptionReq

	optionReq.Options = map[string]string{"replicate.self-heal-daemon": "on"}
	optionReq.Advanced = true

	r.Nil(client.VolumeSet(vol1.Name, optionReq))
	r.Nil(client.SelfHeal(vol1.Name, "index"))
	r.Nil(client.SelfHeal(vol1.Name, "full"))

	// Stop Volume
	r.Nil(client.VolumeStop(vol1.Name), "Volume stop failed")
	// delete volume
	r.Nil(client.VolumeDelete(vol1.Name))
}
