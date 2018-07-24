package restclient

import (
	"fmt"
	"net/http"

	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"
)

// RebalanceStart starts rebalance process for given volume
func (c *Client) RebalanceStart(volname string, option string) error {
	req := rebalanceapi.StartReq{
		Option: option,
	}
	url := fmt.Sprintf("/v1/volumes/%s/rebalance/start", volname)
	return c.post(url, req, http.StatusOK, nil)
}

// RebalanceStop stops rebalance process for given volume
func (c *Client) RebalanceStop(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/rebalance/stop", volname)
	return c.post(url, nil, http.StatusOK, nil)
}
