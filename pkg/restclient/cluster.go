package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// GetClusterOption gets cluster level options
func (c *Client) GetClusterOption() ([]api.ClusterOptionsResp, error) {
	url := fmt.Sprintf("/v1/cluster/options")
	var resp []api.ClusterOptionsResp
	err := c.get(url, nil, http.StatusOK, &resp)
	return resp, err
}

// ClusterOptionSet sets cluster level options
func (c *Client) ClusterOptionSet(req api.ClusterOptionReq) error {
	url := fmt.Sprintf("/v1/cluster/options")
	return c.post(url, req, http.StatusOK, nil)
}
