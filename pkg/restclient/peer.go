package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// PeerAdd adds a peer to the Cluster
func (c *Client) PeerAdd(peerAddReq api.PeerAddReq) (api.PeerAddResp, *http.Response, error) {
	var resp api.PeerAddResp
	httpResp, err := c.post("/v1/peers", peerAddReq, http.StatusCreated, &resp)
	return resp, httpResp, err
}

// PeerRemove removes a peer from the Cluster
func (c *Client) PeerRemove(peerid string) (*http.Response, error) {
	delURL := fmt.Sprintf("/v1/peers/%s", peerid)
	return c.del(delURL, nil, http.StatusNoContent, nil)
}

// GetPeer returns information about a peer
func (c *Client) GetPeer(peerid string) (api.PeerGetResp, *http.Response, error) {
	var peer api.PeerGetResp
	resp, err := c.get("/v1/peers/"+peerid, nil, http.StatusOK, &peer)
	return peer, resp, err
}

// Peers gets list of Gluster Peers
func (c *Client) Peers(filterParams ...map[string]string) (api.PeerListResp, *http.Response, error) {
	var peers api.PeerListResp
	var queryString string
	if len(filterParams) != 0 {
		queryString = getQueryString(filterParams[0])
	}
	resp, err := c.get("/v1/peers"+queryString, nil, http.StatusOK, &peers)
	return peers, resp, err
}
