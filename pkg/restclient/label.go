package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// LabelCreate creates Gluster Label
func (c *Client) LabelCreate(req api.LabelCreateReq) (api.LabelCreateResp, error) {
	var labelinfo api.LabelCreateResp
	err := c.post("/v1/snapshots/labels/create", req, http.StatusCreated, &labelinfo)
	return labelinfo, err
}

//LabelSet will change its values to default.
func (c *Client) LabelSet(req api.LabelSetReq, labelname string) error {
	url := fmt.Sprintf("/v1/snapshots/labels/%s/config", labelname)
	return c.post(url, req, http.StatusOK, nil)
}

//LabelReset will allow to modify its values.
func (c *Client) LabelReset(req api.LabelResetReq, labelname string) error {
	url := fmt.Sprintf("/v1/snapshots/labels/%s/config", labelname)
	return c.del(url, req, http.StatusOK, nil)
}

// LabelList returns list of all labels
func (c *Client) LabelList(labelname string) (api.LabelListResp, error) {
	var labelinfos api.LabelListResp
	err := c.get("/v1/snapshots/labels/list", nil, http.StatusOK, &labelinfos)
	return labelinfos, err
}

// LabelInfo returns information about a label
func (c *Client) LabelInfo(labelname string) (api.LabelGetResp, error) {
	var labelinfo api.LabelGetResp
	var url string
	url = fmt.Sprintf("/v1/snapshots/labels/%s", labelname)
	err := c.get(url, nil, http.StatusOK, &labelinfo)
	return labelinfo, err
}

// LabelDelete will delete Gluster Label and respective lv
func (c *Client) LabelDelete(labelname string) error {
	url := fmt.Sprintf("/v1/snapshots/labels/%s", labelname)
	err := c.del(url, nil, http.StatusNoContent, nil)
	return err
}
