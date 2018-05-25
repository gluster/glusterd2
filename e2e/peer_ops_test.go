package e2e

import (
	"testing"
	"time"

        "github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

func TestAddRemovePeer(t *testing.T) {
	r := require.New(t)

	g1, err := spawnGlusterd("./config/1.toml", true)
	r.Nil(err)
	defer g1.Stop()
	r.True(g1.IsRunning())

	g2, err := spawnGlusterd("./config/2.toml", true)
	r.Nil(err)
	defer g2.Stop()
	r.True(g2.IsRunning())

	g3, err := spawnGlusterd("./config/3.toml", true)
	r.Nil(err)
	defer g3.Stop()
	r.True(g3.IsRunning())

	client := initRestclient(g1.ClientAddress)
        peerAddReq := api.PeerAddReq{
                Addresses: []string{g2.PeerAddress},
                Metadata: map[string]string {
                        "owner": "gd2test",
                },
        }
	_, err2 := client.PeerAdd(peerAddReq)
	r.Nil(err2)

	time.Sleep(6 * time.Second)

	// add peer: ask g1 to add g3 as peer
        peerAddReq = api.PeerAddReq{
                Addresses: []string{g3.PeerAddress},
        }

	_, err3 := client.PeerAdd(peerAddReq)
	r.Nil(err3)

	time.Sleep(6 * time.Second)

	// list and check you have 3 peers in cluster
	peers, err4 := client.Peers()
	r.Nil(err4)
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

	// remove peer: ask g1 to remove g2 as peer
	err5 := client.PeerRemove(g2.PeerID())
	r.Nil(err5)
}
