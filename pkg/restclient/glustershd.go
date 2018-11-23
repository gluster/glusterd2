package restclient

import (
	"errors"
	"fmt"
	"net/http"

	gderrors "github.com/gluster/glusterd2/pkg/errors"
	shdapi "github.com/gluster/glusterd2/plugins/glustershd/api"
)

// SelfHealInfo sends request to heal-info API
func (c *Client) SelfHealInfo(params ...string) ([]shdapi.BrickHealInfo, error) {
	var url string
	if len(params) == 1 {
		url = fmt.Sprintf("/v1/volumes/%s/heal-info", params[0])
	} else if len(params) == 2 {
		url = fmt.Sprintf("/v1/volumes/%s/%s/heal-info", params[0], params[1])
	} else {
		return nil, errors.New("invalid parameters")
	}
	var output []shdapi.BrickHealInfo
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

// SelfHealSplitBrain sends request to start split-brain operations on a volume
func (c *Client) SelfHealSplitBrain(volname, operation string, req shdapi.SplitBrainReq) error {
	var url string
	switch operation {
	case "latest-mtime":
		fallthrough
	case "bigger-file":
		if req.FileName == "" {
			return gderrors.ErrFilenameNotFound
		}
		url = fmt.Sprintf("/v1/volumes/%s/split-brain/%s", volname, operation)

	case "source-brick":
		if req.HostName == "" || req.BrickName == "" {
			return gderrors.ErrHostOrBrickNotFound
		}
		url = fmt.Sprintf("/v1/volumes/%s/split-brain/%s", volname, operation)

	default:
		return gderrors.ErrInvalidSplitBrainOp
	}
	return c.post(url, req, http.StatusOK, nil)
}
