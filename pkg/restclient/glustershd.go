package restclient

import (
	"errors"
	"fmt"
	"net/http"

	glustershdapi "github.com/gluster/glusterd2/plugins/glustershd/api"
)

// SelfHealInfo sends request to heal-info API
func (c *Client) SelfHealInfo(params ...string) ([]glustershdapi.BrickHealInfo, error) {
	var url string
	if len(params) == 1 {
		url = fmt.Sprintf("/v1/volumes/%s/heal-info", params[0])
	} else if len(params) == 2 {
		url = fmt.Sprintf("/v1/volumes/%s/%s/heal-info", params[0], params[1])
	} else {
		return nil, errors.New("invalid parameters")
	}
	var output []glustershdapi.BrickHealInfo
	err := c.get(url, nil, http.StatusOK, &output)
	return output, err
}

// SelfHeal sends request to start the heal process on the specified volname
func (c *Client) SelfHeal(volname string, healType string) error {
	var url string
	switch healType {
	case "index":
		url = fmt.Sprintf("/v1/volumes/%s/heal", volname)
	case "full":
		url = fmt.Sprintf("/v1/volumes/%s/heal?type=%s", volname, healType)
	default:
		return errors.New("invalid parameters")
	}

	return c.post(url, nil, http.StatusOK, nil)
}
