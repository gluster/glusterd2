package e2e

import (
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

func TestAddRemovePeer(t *testing.T) {
	r := require.New(t)

	// set up a cluster w/o glusterd instances for dependencies
	tc, err := setupCluster(t)
	r.NoError(err)
	defer teardownCluster(tc)

	g1, err := spawnGlusterd(t, "./config/1.toml", true)
	r.Nil(err)
	defer g1.Stop()
	r.True(g1.IsRunning())

	g2, err := spawnGlusterd(t, "./config/2.toml", true)
	r.Nil(err)
	defer g2.Stop()
	r.True(g2.IsRunning())

	g3, err := spawnGlusterd(t, "./config/3.toml", true)
	r.Nil(err)
	defer g3.Stop()
	r.True(g3.IsRunning())

	client, err := initRestclient(g1)
	r.Nil(err)
	r.NotNil(client)

	peerAddReq := api.PeerAddReq{
		Addresses: []string{g2.PeerAddress},
		Metadata: map[string]string{
			"owner": "gd2test",
		},
	}
	_, err = client.PeerAdd(peerAddReq)
	r.Nil(err)

	// add peer: ask g1 to add g3 as peer
	peerAddReq = api.PeerAddReq{
		Addresses: []string{g3.PeerAddress},
	}

	peerinfo, err := client.PeerAdd(peerAddReq)
	r.Nil(err)

	_, err = client.GetPeer(peerinfo.ID.String())
	r.Nil(err)

	// list and check you have 3 peers in cluster
	peers, err := client.Peers()
	r.Nil(err)
	r.Len(peers, 3)

	var matchingQueries []map[string]string
	var nonMatchingQueries []map[string]string

	matchingQueries = append(matchingQueries, map[string]string{
		"key":   "owner",
		"value": "gd2test",
	})
	matchingQueries = append(matchingQueries, map[string]string{
		"key": "owner",
	})
	matchingQueries = append(matchingQueries, map[string]string{
		"value": "gd2test",
	})
	for _, filter := range matchingQueries {
		peers, err := client.Peers(filter)
		r.Nil(err)
		r.Len(peers, 1)
	}

	nonMatchingQueries = append(nonMatchingQueries, map[string]string{
		"key":   "owner",
		"value": "gd2-test",
	})
	nonMatchingQueries = append(nonMatchingQueries, map[string]string{
		"key": "owners",
	})
	nonMatchingQueries = append(nonMatchingQueries, map[string]string{
		"value": "gd2tests",
	})
	for _, filter := range nonMatchingQueries {
		peers, err := client.Peers(filter)
		r.Nil(err)
		r.Len(peers, 0)
	}

	devicesDir := testTempDir(t, "devices")
	r.Nil(prepareLoopDevice(devicesDir+"/gluster_dev1.img", "1", "250M"))
	_, err = client.DeviceAdd(g2.PeerID(), "/dev/gluster_loop1")
	r.Nil(err)

	dev, err := client.DeviceList(g2.PeerID(), "/dev/gluster_loop1")
	r.Nil(err)
	r.Equal(dev[0].Device, "/dev/gluster_loop1")

	err = client.PeerRemove(g2.PeerID())
	r.NotNil(err)

	err = client.PeerRemove(g3.PeerID())
	r.Nil(err)

	// Device Cleanup
	r.Nil(loopDevicesCleanup(t))
}
