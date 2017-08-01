package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gluster/glusterd2/peer"
)

func TestAddRemovePeer(t *testing.T) {
	r := require.New(t)

	g1, err := spawnGlusterd("./config/1.yaml", true)
	r.Nil(err)
	defer g1.Stop()
	defer g1.EraseWorkdir()
	r.True(g1.IsRunning())

	g2, err := spawnGlusterd("./config/2.yaml", true)
	r.Nil(err)
	defer g2.Stop()
	defer g2.EraseWorkdir()
	r.True(g2.IsRunning())

	g3, err := spawnGlusterd("./config/3.yaml")
	r.Nil(err)
	defer g3.Stop()
	defer g3.EraseWorkdir()
	r.True(g3.IsRunning())

	// add peer: ask g1 to add g2 as peer
	reqBody := strings.NewReader(fmt.Sprintf(`{"addresses": ["%s"]}`, g2.PeerAddress))
	resp, err := http.Post("http://"+g1.ClientAddress+"/v1/peers", "application/json", reqBody)
	r.Nil(err)
	defer resp.Body.Close()
	r.Equal(http.StatusCreated, resp.StatusCode)

	time.Sleep(5 * time.Second)

	// add peer: ask g1 to add g3 as peer
	reqBody = strings.NewReader(fmt.Sprintf(`{"addresses": ["%s"]}`, g3.PeerAddress))
	resp, err = http.Post("http://"+g1.ClientAddress+"/v1/peers", "application/json", reqBody)
	r.Nil(err)
	defer resp.Body.Close()
	r.Equal(http.StatusCreated, resp.StatusCode)

	time.Sleep(5 * time.Second)

	// list and check you have 2 peers in cluster
	resp, err = http.Get("http://" + g1.ClientAddress + "/v1/peers")
	r.Nil(err)
	defer resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	r.Nil(err)
	var peers []peer.Peer
	err = json.Unmarshal(data, &peers)
	r.Nil(err)
	r.Len(peers, 3)

	// remove peer: ask g1 to remove g2 as peer
	delURL := fmt.Sprintf("http://%s/v1/peers/%s", g1.ClientAddress, g2.PeerID())
	req, err := http.NewRequest("DELETE", delURL, nil)
	r.Nil(err)
	resp, err = http.DefaultClient.Do(req)
	r.Nil(err)
	defer resp.Body.Close()
	r.Equal(http.StatusNoContent, resp.StatusCode)
}
