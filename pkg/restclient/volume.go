package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// VolumeCreate creates Gluster Volume
func (c *Client) VolumeCreate(req api.VolCreateReq) (api.VolumeCreateResp, error) {
	var vol api.VolumeCreateResp
	err := c.post("/v1/volumes", req, http.StatusCreated, &vol)
	return vol, err
}

// Volumes returns list of all volumes
func (c *Client) Volumes(volname string) (api.VolumeListResp, error) {
	if volname == "" {
		var vols api.VolumeListResp
		url := fmt.Sprintf("/v1/volumes")
		err := c.get(url, nil, http.StatusOK, &vols)
		return vols, err
	}
	var vol api.VolumeGetResp
	url := fmt.Sprintf("/v1/volumes/%s", volname)
	err := c.get(url, nil, http.StatusOK, &vol)
	return []api.VolumeGetResp{vol}, err
}

// BricksStatus returns the status of bricks that form a Gluster volume
func (c *Client) BricksStatus(volname string) (api.BricksStatusResp, error) {
	url := fmt.Sprintf("/v1/volumes/%s/bricks", volname)
	var resp api.BricksStatusResp
	err := c.get(url, nil, http.StatusOK, &resp)
	return resp, err
}

// VolumeStatus returns the status of a Gluster volume
func (c *Client) VolumeStatus(volname string) (api.VolumeStatusResp, error) {
	url := fmt.Sprintf("/v1/volumes/%s/status", volname)
	var volStatus api.VolumeStatusResp
	err := c.get(url, nil, http.StatusOK, &volStatus)
	return volStatus, err
}

// VolumeStart starts a Gluster Volume
func (c *Client) VolumeStart(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/start", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// VolumeStop stops a Gluster Volume
func (c *Client) VolumeStop(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/stop", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// VolumeDelete deletes a Gluster Volume
func (c *Client) VolumeDelete(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s", volname)
	return c.del(url, nil, http.StatusOK, nil)
}

// VolumeSet sets an option for a Gluster Volume
func (c *Client) VolumeSet(volname string, req api.VolOptionReq) error {
	url := fmt.Sprintf("/v1/volumes/%s/options", volname)
	err := c.post(url, req, http.StatusOK, nil)
	return err
}

// GlobalOptionSet sets cluster level options
func (c *Client) GlobalOptionSet(req api.GlobalOptionReq) error {
	url := fmt.Sprintf("/v1/cluster/options")
	return c.post(url, req, http.StatusOK, nil)
}

// VolumeExpand expands a Gluster Volume
func (c *Client) VolumeExpand(volname string, req api.VolExpandReq) (api.VolumeExpandResp, error) {
	var vol api.VolumeExpandResp
	url := fmt.Sprintf("/v1/volumes/%s/expand", volname)
	err := c.post(url, req, http.StatusOK, &vol)
	return vol, err
}

// OptionGroupCreate creates a new option group
func (c *Client) OptionGroupCreate(req api.OptionGroupReq) error {
	return c.post("/v1/volumes/options-group", req, http.StatusCreated, nil)
}

// OptionGroupList returns a list of all option groups
func (c *Client) OptionGroupList() (api.OptionGroupListResp, error) {
	var l api.OptionGroupListResp
	url := fmt.Sprintf("/v1/volumes/options-group")
	err := c.get(url, nil, http.StatusOK, &l)
	return l, err
}

// OptionGroupDelete deletes the specified option group
func (c *Client) OptionGroupDelete(group string) error {
	url := fmt.Sprintf("/v1/volumes/options-group/%s", group)
	return c.del(url, nil, http.StatusOK, nil)
}
