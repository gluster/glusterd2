package restclient

import (
	"fmt"
	"net/http"

	bitrotapi "github.com/gluster/glusterd2/plugins/bitrot/api"
)

// BitrotEnable enables bitrot for a volume
func (c *Client) BitrotEnable(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/bitrot/enable", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// BitrotDisable disables bitrot for a volume
func (c *Client) BitrotDisable(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/bitrot/disable", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// BitrotScrubOndemand starts bitrot scrubber on demand for a volume
func (c *Client) BitrotScrubOndemand(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/bitrot/scrubondemand", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// BitrotScrubStatus returns bitrot scrub status of a volume
func (c *Client) BitrotScrubStatus(volname string) (bitrotapi.ScrubStatus, error) {
	var scrubStatus bitrotapi.ScrubStatus
	url := fmt.Sprintf("/v1/volumes/%s/bitrot/scrubstatus", volname)
	err := c.get(url, nil, http.StatusOK, &scrubStatus)
	return scrubStatus, err
}
