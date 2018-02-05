package restclient

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// PeerProbe adds a peer to the Cluster
func (c *Client) PeerProbe(host string) (api.PeerAddResp, error) {

	peerAddReq := api.PeerAddReq{
		Addresses: []string{host},
	}

	var resp api.PeerAddResp
	err := c.post("/v1/peers", peerAddReq, http.StatusCreated, &resp)
	return resp, err
}

// PeerDetach detaches a peer from the Cluster
func (c *Client) PeerDetach(host string) error {
	// Get Peers list to find Peer ID
	peers, err := c.Peers()
	if err != nil {
		return err
	}

	peerID := ""

	// Find Peer ID using available information
	for _, p := range peers {
		for _, h := range p.PeerAddresses {
			if h == host {
				peerID = p.ID.String()
				break
			}
		}
		// If already got Peer ID
		if peerID != "" {
			break
		}
	}

	if peerID == "" {
		return errors.New("Unable to find Peer ID")
	}

	return c.PeerDetachByID(peerID)
}

// PeerDetachByID detaches a peer from the Cluster
func (c *Client) PeerDetachByID(peerid string) error {
	delURL := fmt.Sprintf("/v1/peers/%s", peerid)
	return c.del(delURL, nil, http.StatusNoContent, nil)
}

// Peers gets list of Gluster Peers
func (c *Client) Peers() (api.PeerListResp, error) {
	var peers api.PeerListResp
	err := c.get("/v1/peers", nil, http.StatusOK, &peers)
	return peers, err
}
