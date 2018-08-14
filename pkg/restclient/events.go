package restclient

import (
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"
)

// WebhookAdd registers webhook to listen to Gluster Events
func (c *Client) WebhookAdd(url string, token string, secret string) (*http.Response, error) {
	req := &eventsapi.Webhook{
		URL:    url,
		Token:  token,
		Secret: secret,
	}
	return c.post("/v1/events/webhook", req, http.StatusOK, nil)
}

// WebhookDelete deletes the webhook
func (c *Client) WebhookDelete(url string) (*http.Response, error) {
	req := &eventsapi.WebhookDel{
		URL: url,
	}

	return c.del("/v1/events/webhook", req, http.StatusNoContent, nil)
}

// Webhooks returns the list of Webhooks listening to Gluster Events
func (c *Client) Webhooks() (eventsapi.WebhookList, *http.Response, error) {
	var resp eventsapi.WebhookList
	httpResp, err := c.get("/v1/events/webhook", nil, http.StatusOK, &resp)
	return resp, httpResp, err
}

// ListEvents returns the list of Gluster Events
func (c *Client) ListEvents() ([]*api.Event, *http.Response, error) {
	var resp []*api.Event
	httpResp, err := c.get("/v1/events", nil, http.StatusOK, &resp)
	return resp, httpResp, err
}

// WebhookTest tests connection between peers and specified URL
func (c *Client) WebhookTest(url string, token string, secret string) (*http.Response, error) {
	req := &eventsapi.Webhook{
		URL:    url,
		Token:  token,
		Secret: secret,
	}
	return c.post("/v1/events/webhook/test", req, http.StatusOK, nil)
}
