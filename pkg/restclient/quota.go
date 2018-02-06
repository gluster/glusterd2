package restclient

import (
	"fmt"
	"net/http"
)

// QuotaEnable starts a Gluster Volume
func (c *Client) QuotaEnable(volname string) error {
	url := fmt.Sprintf("/v1/quota/%s", volname)
	return c.post(url, nil, http.StatusOK, nil)
}
