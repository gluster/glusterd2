package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// PeerProbe adds a peer to the Cluster
func (c *Client) PeerProbe(host string) error {
	peerAddReq := api.PeerAddReq{
		Addresses: []string{host},
	}
	return c.action("POST", "/v1/peers", peerAddReq, http.StatusCreated)
}

// PeerDetach detaches a peer from the Cluster
func (c *Client) PeerDetach(host string) error {
	delURL := fmt.Sprintf("/v1/peers/%s", host)
	return c.action("DELETE", delURL, nil, http.StatusNoContent)
}
