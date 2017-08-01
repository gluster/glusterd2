package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRestart tests that data persists after a GD2 restart
func TestRestart(t *testing.T) {
	r := require.New(t)

	gd, err := spawnGlusterd("./config/1.yaml", true)
	r.Nil(err)
	r.True(gd.IsRunning())

	reqT := `
{
  "name": "vol1",
  "bricks": ["%s:%s"],
  "force": true
}
	`
	dir, err := ioutil.TempDir("", "")
	r.Nil(err)
	defer os.RemoveAll(dir)

	req := fmt.Sprintf(reqT, gd.PeerAddress, dir)

	resp, err := http.Post(fmt.Sprintf("http://%s/v1/volumes", gd.ClientAddress), "application/json", strings.NewReader(req))
	r.Nil(err)
	defer resp.Body.Close()
	r.Equal(http.StatusCreated, resp.StatusCode)

	r.Len(getVols(gd, r), 1)

	r.Nil(gd.Stop())

	gd, err = spawnGlusterd("./config/1.yaml", false)
	r.Nil(err)
	r.True(gd.IsRunning())

	r.Len(getVols(gd, r), 1)

	r.Nil(gd.Stop())
}

func getVols(gd *gdProcess, r *require.Assertions) map[string]string {
	resp, err := http.Get(fmt.Sprintf("http://%s/v1/volumes", gd.ClientAddress))
	r.Nil(err)
	r.Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	vols := make(map[string]string)
	data, err := ioutil.ReadAll(resp.Body)
	r.Nil(err)
	err = json.Unmarshal(data, &vols)
	r.Nil(err)

	return vols
}
