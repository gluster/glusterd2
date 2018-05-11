package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// PeerAdd adds a peer to the Cluster
func (c *Client) PeerAdd(host string) (api.PeerAddResp, error) {

	peerAddReq := api.PeerAddReq{
		Addresses: []string{host},
	}

	var resp api.PeerAddResp
	err := c.post("/v1/peers", peerAddReq, http.StatusCreated, &resp)
	return resp, err
}

// PeerRemove removes a peer from the Cluster
func (c *Client) PeerRemove(peerid string) error {
	delURL := fmt.Sprintf("/v1/peers/%s", peerid)
	return c.del(delURL, nil, http.StatusNoContent, nil)
}

// Peers gets list of Gluster Peers
func (c *Client) Peers() (api.PeerListResp, error) {
	var peers api.PeerListResp
	err := c.get("/v1/peers", nil, http.StatusOK, &peers)
	return peers, err
}
