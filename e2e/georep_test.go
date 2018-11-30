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

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	brickDir := testTempDir(t, "bricks")
	defer os.RemoveAll(brickDir)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	volname := formatVolName(t.Name())
	reqVol := api.VolCreateReq{
		Name: volname,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
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

	err = client.VolumeStart(vol1.Name, false)
	r.Nil(err)

	volname2 := "testvol2"
	reqVol = api.VolCreateReq{
		Name: volname2,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[2]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Force: true,
	}
	vol2, err := client.VolumeCreate(reqVol)
	r.Nil(err)

	reqGeorep := georepapi.GeorepCreateReq{
		MasterVol: volname,
		RemoteVol: volname2,
		RemoteHosts: []georepapi.GeorepRemoteHostReq{
			{PeerID: tc.gds[1].PeerID(), Hostname: tc.gds[1].PeerAddress},
		},
	}

	masterVolID := vol1.ID.String()
	remoteVolID := vol2.ID.String()

	_, err = client.GeorepCreate(masterVolID, remoteVolID, reqGeorep)
	r.Nil(err)

	//generate ssh keys
	_, err = client.GeorepSSHKeysGenerate(volname)
	r.Nil(err)

	//get ssh keys
	sshKeys, err := client.GeorepSSHKeys(volname)
	r.Nil(err)

	//push ssh keys
	err = client.GeorepSSHKeysPush(volname, sshKeys)
	r.Nil(err)

	//start geo-rep session
	_, err = client.GeorepStart(masterVolID, remoteVolID, false)
	r.Nil(err)

	//set geo-rep options
	opt := make(map[string]string)
	opt["gluster-log-level"] = "INFO"
	opt["changelog-log-level"] = "ERROR"
	err = client.GeorepSet(masterVolID, remoteVolID, opt)
	r.Nil(err)

	//get geo-rep options
	_, err = client.GeorepGet(masterVolID, remoteVolID)
	r.Nil(err)

	//reset geo-rep options
	err = client.GeorepReset(masterVolID, remoteVolID, []string{"gluster-log-level", "changelog-log-level"})
	r.Nil(err)
	//pause geo-rep session
	_, err = client.GeorepPause(masterVolID, remoteVolID, false)
	r.Nil(err)

	//resume geo-rep session
	_, err = client.GeorepResume(masterVolID, remoteVolID, false)
	r.Nil(err)

	//stop geo-rep session
	_, err = client.GeorepStop(masterVolID, remoteVolID, false)
	r.Nil(err)

	//get status of geo-rep session
	_, err = client.GeorepStatus(masterVolID, remoteVolID)
	r.Nil(err)

	//gets status of geo-rep sessions
	_, err = client.GeorepStatus("", "")
	r.Nil(err)

	// delete geo-rep session
	err = client.GeorepDelete(masterVolID, remoteVolID, false)
	r.Nil(err)

	// stop volume
	err = client.VolumeStop(volname)
	r.Nil(err)

	// delete volume
	err = client.VolumeDelete(volname)
	r.Nil(err)

	// delete volume
	err = client.VolumeDelete(volname2)
	r.Nil(err)
}
