package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// It'll be a good idea to create a separate subpackage in glusterd2 source
// with these request types so that clients can also make use of them.
type volCreateReq struct {
	Name      string   `json:"name"`
	Transport string   `json:"transport,omitempty"`
	Replica   int      `json:"replica,omitempty"`
	Bricks    []string `json:"bricks"`
	Force     bool     `json:"force,omitempty"`
}

func TestVolumeCreateDelete(t *testing.T) {
	assert := assert.New(t)

	gds, err := setupCluster("./config/1.yaml", "./config/2.yaml")
	assert.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir("", "TestVolumeCreateDelete")
	defer os.RemoveAll(brickDir)

	var brickPaths []string
	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		assert.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	// create 2x2 dist-rep volume
	volname := "testvol"
	createReq := volCreateReq{
		Name:    volname,
		Replica: 2,
		Bricks: []string{
			gds[0].PeerAddress + ":" + brickPaths[0],
			gds[1].PeerAddress + ":" + brickPaths[1],
			gds[0].PeerAddress + ":" + brickPaths[2],
			gds[1].PeerAddress + ":" + brickPaths[3]},
		Force: true,
	}
	reqBody, err := json.Marshal(createReq)
	assert.Nil(err)

	volCreateURL := fmt.Sprintf("http://%s/v1/volumes", gds[0].ClientAddress)
	resp, err := http.Post(volCreateURL, "application/json", strings.NewReader(string(reqBody)))
	assert.Nil(err)
	defer resp.Body.Close()
	assert.Equal(resp.StatusCode, 201)

	// delete volume
	volDelURL := fmt.Sprintf("http://%s/v1/volumes/%s", gds[0].ClientAddress, volname)
	delReq, err := http.NewRequest("DELETE", volDelURL, nil)
	assert.Nil(err)
	resp, err = http.DefaultClient.Do(delReq)
	assert.Nil(err)
	defer resp.Body.Close()
	assert.Equal(resp.StatusCode, 200)
}
