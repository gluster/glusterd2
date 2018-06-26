package e2e

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

var (
	webhookURL string
)

// TestWebhook test webhooks and events
func TestWebhook(t *testing.T) {
	var err error

	r := require.New(t)

	gds, err = setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	client = initRestclient(gds[0].ClientAddress)

	tmpDir, err = ioutil.TempDir(baseWorkdir, t.Name())
	r.Nil(err)
	t.Logf("Using temp dir: %s", tmpDir)

	t.Run("Register-webhook", testAddWebhook)
	t.Run("List-webhook", testGetWebhook)
	t.Run("Delete-webhook", testDeleteWebhook)
	t.Run("List-gluster-events", testEvents)

}

func testAddWebhook(t *testing.T) {
	r := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(hw http.ResponseWriter, hr *http.Request) {
		hw.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	webhookURL = ts.URL
	//create webhook
	r.Nil(client.WebhookAdd(webhookURL, "", ""))

	volumeName := formatVolName(t.Name())
	var brickPaths []string

	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(tmpDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	// create volume
	createReq := api.VolCreateReq{
		Name: volumeName,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Force: true,
	}

	_, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeDelete(volumeName))
}

func testGetWebhook(t *testing.T) {
	r := require.New(t)

	webhooks, err := client.Webhooks()
	r.Nil(err)
	r.Equal(webhooks[0], webhookURL)
}

func testDeleteWebhook(t *testing.T) {
	r := require.New(t)
	//delete webhook
	r.Nil(client.WebhookDelete(webhookURL))

}

func testEvents(t *testing.T) {
	r := require.New(t)

	events, err := client.ListEvents()
	r.Nil(err)
	r.NotEmpty(events)
}
