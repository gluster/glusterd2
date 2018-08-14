package restclient

import (
	"errors"
	"fmt"
	"net/http"

	glustershdapi "github.com/gluster/glusterd2/plugins/glustershd/api"
)

// SelfHealInfo sends request to heal-info API
func (c *Client) SelfHealInfo(params ...string) ([]glustershdapi.BrickHealInfo, *http.Response, error) {
	var url string
	if len(params) == 1 {
		url = fmt.Sprintf("/v1/volumes/%s/heal-info", params[0])
	} else if len(params) == 2 {
		url = fmt.Sprintf("/v1/volumes/%s/%s/heal-info", params[0], params[1])
	} else {
		return nil, nil, errors.New("invalid parameters")
	}
	var output []glustershdapi.BrickHealInfo
	resp, err := c.get(url, nil, http.StatusOK, &output)
	return output, resp, err
}

// SelfHeal sends request to start the heal process on the specified volname
func (c *Client) SelfHeal(volname string, healType string) (*http.Response, error) {
	var url string
	switch healType {
	case "index":
		url = fmt.Sprintf("/v1/volumes/%s/heal", volname)
	case "full":
		url = fmt.Sprintf("/v1/volumes/%s/heal?type=%s", volname, healType)
	default:
		return nil, errors.New("invalid parameters")
	}

	return c.post(url, nil, http.StatusOK, nil)
}
