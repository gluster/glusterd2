package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRemovePeer(t *testing.T) {
	assert := assert.New(t)

	g1, err := spawnGlusterd("./config/1.yaml")
	assert.Nil(err)
	defer g1.Stop()
	defer g1.EraseWorkdir()
	assert.True(g1.IsRunning())

	g2, err := spawnGlusterd("./config/2.yaml")
	assert.Nil(err)
	defer g2.Stop()
	defer g2.EraseWorkdir()
	assert.True(g2.IsRunning())

	// add peer: ask g1 to add g2 as peer
	reqBody := strings.NewReader(fmt.Sprintf(`{"addresses": ["%s"]}`, g2.PeerAddress))
	resp, err := http.Post("http://"+g1.ClientAddress+"/v1/peers", "application/json", reqBody)
	assert.Nil(err)
	defer resp.Body.Close()
	assert.Equal(resp.StatusCode, 201)

	// remove peer: ask g1 to remove g2 as peer
	delURL := fmt.Sprintf("http://%s/v1/peers/%s", g1.ClientAddress, g2.PeerID())
	req, err := http.NewRequest("DELETE", delURL, nil)
	assert.Nil(err)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(err)
	defer resp.Body.Close()
	assert.Equal(resp.StatusCode, 204)
}
