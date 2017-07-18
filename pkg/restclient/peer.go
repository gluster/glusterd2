package restclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"
)

// PeerProbe adds a peer to the Cluster
func (c *RESTClient) PeerProbe(host string) error {
	peerAddReq := api.PeerAddReq{
		Addresses: []string{host},
	}
	reqBody, err := json.Marshal(peerAddReq)
	if err != nil{
		return err
	}
	return httpRESTAction("POST", c.baseURL+"/v1/peers", strings.NewReader(string(reqBody)), 201)
}

// PeerDetach detaches a peer from the Cluster
func (c *RESTClient) PeerDetach(host string) error {
	delURL := fmt.Sprintf(c.baseURL+"/v1/peers/%s", host)
	return httpRESTAction("DELETE", delURL, nil, 204)
}
