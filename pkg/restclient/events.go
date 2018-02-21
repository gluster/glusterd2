package restclient

import (
	"net/http"

	eventsapi "github.com/gluster/glusterd2/plugins/events/api"
)

// WebhookAdd registers webhook to listen to Gluster Events
func (c *Client) WebhookAdd(url string, token string, secret string) error {
	req := &eventsapi.Webhook{
		URL:    url,
		Token:  token,
		Secret: secret,
	}
	return c.post("/v1/events/webhook", req, http.StatusOK, nil)
}

// WebhookDelete deletes the webhook
func (c *Client) WebhookDelete(url string) error {
	req := &eventsapi.Webhook{
		URL: url,
	}

	return c.del("/v1/events/webhook", req, http.StatusOK, nil)
}

// Webhooks returns the list of Webhooks listening to Gluster Events
func (c *Client) Webhooks() (eventsapi.WebhookList, error) {
	var resp eventsapi.WebhookList
	err := c.get("/v1/events/webhook", nil, http.StatusOK, &resp)
	return resp, err
}
