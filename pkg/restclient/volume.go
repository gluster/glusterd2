package restclient

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gluster/glusterd2/pkg/api"
)

// metadataFilter is a filter type
type metadataFilter uint32

// GetVolumes Filter Types
const (
	noKeyAndValue metadataFilter = iota
	onlyKey
	onlyValue
	keyAndValue
)

// VolumeCreate creates Gluster Volume
func (c *Client) VolumeCreate(req api.VolCreateReq) (api.VolumeCreateResp, error) {
	var vol api.VolumeCreateResp
	err := c.post("/v1/volumes", req, http.StatusCreated, &vol)
	return vol, err
}

// getFilterType return the filter type for volume list/info
func getFilterType(filterParams map[string]string) metadataFilter {
	_, key := filterParams["key"]
	_, value := filterParams["value"]
	if key && !value {
		return onlyKey
	} else if value && !key {
		return onlyValue
	} else if value && key {
		return keyAndValue
	}
	return noKeyAndValue
}

// getQueryString returns the query string for filtering volumes
func getQueryString(filterParam map[string]string) string {
	filterType := getFilterType(filterParam)
	var queryString string
	switch filterType {
	case onlyKey:
		queryString = fmt.Sprintf("?key=%s", url.QueryEscape(filterParam["key"]))
	case onlyValue:
		queryString = fmt.Sprintf("?value=%s", url.QueryEscape(filterParam["value"]))
	case keyAndValue:
		queryString = fmt.Sprintf("?key=%s&value=%s", url.QueryEscape(filterParam["key"]), url.QueryEscape(filterParam["value"]))
	}
	return queryString
}

// Volumes returns list of all volumes
func (c *Client) Volumes(volname string, filterParams ...map[string]string) (api.VolumeListResp, error) {
	if volname == "" {
		var vols api.VolumeListResp
		var queryString string
		if len(filterParams) > 0 {
			queryString = getQueryString(filterParams[0])
		}
		url := fmt.Sprintf("/v1/volumes%s", queryString)
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
func (c *Client) VolumeStart(volname string, force bool) error {
	req := api.VolumeStartReq{
		ForceStartBricks: force,
	}
	url := fmt.Sprintf("/v1/volumes/%s/start", volname)
	return c.post(url, req, http.StatusOK, nil)
}

// VolumeStop stops a Gluster Volume
func (c *Client) VolumeStop(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s/stop", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// VolumeDelete deletes a Gluster Volume
func (c *Client) VolumeDelete(volname string) error {
	url := fmt.Sprintf("/v1/volumes/%s", volname)
	return c.del(url, nil, http.StatusNoContent, nil)
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

// VolumeGet gets volume options for a Gluster Volume
func (c *Client) VolumeGet(volname string, optname string) (api.VolumeOptionsGetResp, error) {
	if optname == "all" {
		var opts api.VolumeOptionsGetResp
		url := fmt.Sprintf("/v1/volumes/%s/options", volname)
		err := c.get(url, nil, http.StatusOK, &opts)
		return opts, err
	}
	var opt api.VolumeOptionGetResp
	url := fmt.Sprintf("/v1/volumes/%s/options/%s", volname, optname)
	err := c.get(url, nil, http.StatusOK, &opt)
	return []api.VolumeOptionGetResp{opt}, err
}

// VolumeExpand expands a Gluster Volume
func (c *Client) VolumeExpand(volname string, req api.VolExpandReq) (api.VolumeExpandResp, error) {
	var vol api.VolumeExpandResp
	url := fmt.Sprintf("/v1/volumes/%s/expand", volname)
	err := c.post(url, req, http.StatusOK, &vol)
	return vol, err
}

// VolumeStatedump takes statedump of various daemons
func (c *Client) VolumeStatedump(volname string, req api.VolStatedumpReq) error {
	url := fmt.Sprintf("/v1/volumes/%s/statedump", volname)
	return c.post(url, req, http.StatusOK, nil)
}

// OptionGroupCreate creates a new option group
func (c *Client) OptionGroupCreate(req api.OptionGroupReq) error {
	return c.post("/v1/volumes/options-group", req, http.StatusOK, nil)
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
	return c.del(url, nil, http.StatusNoContent, nil)
}

// EditVolume edits the specified keys in volinfo of a volume
func (c *Client) EditVolume(volname string, req api.VolEditReq) (api.VolumeEditResp, error) {
	var resp api.VolumeEditResp
	url := fmt.Sprintf("/v1/volumes/%s/edit", volname)
	err := c.post(url, req, http.StatusOK, &resp)
	return resp, err
}

// VolumeReset resets volume options to their default values
func (c *Client) VolumeReset(volname string, req api.VolOptionResetReq) error {
	url := fmt.Sprintf("/v1/volumes/%s/options", volname)
	return c.del(url, req, http.StatusOK, nil)
}
