package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// PeerAdd adds a peer to the Cluster
func (c *Client) PeerAdd(peerAddReq api.PeerAddReq) (api.PeerAddResp, error) {
	var resp api.PeerAddResp
	err := c.post("/v1/peers", peerAddReq, http.StatusCreated, &resp)
	return resp, err
}

// PeerRemove removes a peer from the Cluster
func (c *Client) PeerRemove(peerid string) error {
	delURL := fmt.Sprintf("/v1/peers/%s", peerid)
	return c.del(delURL, nil, http.StatusNoContent, nil)
}

// GetPeer returns information about a peer
func (c *Client) GetPeer(peerid string) (api.PeerGetResp, error) {
	var peer api.PeerGetResp
	err := c.get("/v1/peers/"+peerid, nil, http.StatusOK, &peer)
	return peer, err
}

// Peers gets list of Gluster Peers
func (c *Client) Peers(filterParams ...map[string]string) (api.PeerListResp, error) {
	var peers api.PeerListResp
	var queryString string
	if len(filterParams) != 0 {
		queryString = getQueryString(filterParams[0])
	}
	err := c.get("/v1/peers"+queryString, nil, http.StatusOK, &peers)
	return peers, err
}
