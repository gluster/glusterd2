package restclient

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/glustershd/api"
)

// SelfHealInfo sends request to heal-info API
func (c *Client) SelfHealInfo(params ...string) ([]api.BrickHealInfo, error) {
	var url string
	if len(params) == 1 {
		url = fmt.Sprintf("/v1/volumes/%s/heal-info", params[0])
	} else if len(params) == 2 {
		url = fmt.Sprintf("/v1/volumes/%s/%s/heal-info", params[0], params[1])
	} else {
		return nil, errors.New("invalid parameters")
	}
	var output []api.BrickHealInfo
	err := c.get(url, nil, http.StatusOK, &output)
	return output, err
}
