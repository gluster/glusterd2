package restclient

import (
	"fmt"
	"net/http"
)

// GlusterShdEnable sends request to api to  enable shd.
func (c *Client) GlusterShdEnable(volname string) error {

	url := fmt.Sprintf("/v1/volumes/%s/heal/enable", volname)
	err := c.post(url, nil, http.StatusOK, nil)
	return err
}

// GlusterShdDisable sends request to api to disable shd.
func (c *Client) GlusterShdDisable(volname string) error {

	url := fmt.Sprintf("/v1/volumes/%s/heal/disable", volname)
	err := c.post(url, nil, http.StatusOK, nil)
	return err
}
