package e2e

import (
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

	tc, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client = initRestclient(tc.gds[0])

	t.Run("Register-webhook", tc.wrap(testAddWebhook))
	t.Run("List-webhook", testGetWebhook)
	t.Run("Delete-webhook", testDeleteWebhook)
	t.Run("List-gluster-events", testEvents)
	t.Run("Webhook-connection", testwebhookconnection)

}

func testAddWebhook(t *testing.T, tc *testCluster) {
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
		brickPath := testTempDir(t, "brick")
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
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: tc.gds[1].PeerID(), Path: brickPaths[3]},
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

func testwebhookconnection(t *testing.T) {
	r := require.New(t)
	var c int

	ts := httptest.NewServer(http.HandlerFunc(func(hw http.ResponseWriter, hr *http.Request) {
		c++
		hw.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	webhookURL = ts.URL
	//test webhook connection
	r.Nil(client.WebhookTest(webhookURL, "", ""))

	peers, err := client.Peers()
	r.Nil(err)

	if c != len(peers) {
		r.Fail("failed to test webhook connection")
	}
}
