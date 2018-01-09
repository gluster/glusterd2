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

// VolumeExpand expands a Gluster Volume
func (c *Client) VolumeExpand(volname string, req api.VolExpandReq) (api.VolumeExpandResp, error) {
	var vol api.VolumeExpandResp
	url := fmt.Sprintf("/v1/volumes/%s/expand", volname)
	err := c.post(url, req, http.StatusOK, &vol)
	return vol, err
}

// ProfileCreate creates a new profile
func (c *Client) ProfileCreate(req api.ProfileCreateReq) error {
	return c.post("/v1/volumes/options", req, http.StatusCreated, nil)
}

// ProfileTunables returns the list of tunables that is part of a profile
func (c *Client) ProfileTunables(profilename string) (api.ProfileTunablesResp, error) {
	var tunables api.ProfileTunablesResp
	url := fmt.Sprintf("/v1/volumes/options/%s", profilename)
	err := c.get(url, nil, http.StatusOK, &tunables)
	return tunables, err
}
