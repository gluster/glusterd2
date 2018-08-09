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
func (c *Client) VolumeCreate(req api.VolCreateReq) (api.VolumeCreateResp, *http.Response, error) {
	var vol api.VolumeCreateResp
	resp, err := c.post("/v1/volumes", req, http.StatusCreated, &vol)
	return vol, resp, err
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
func (c *Client) Volumes(volname string, filterParams ...map[string]string) (api.VolumeListResp, *http.Response, error) {
	if volname == "" {
		var vols api.VolumeListResp
		var queryString string
		if len(filterParams) > 0 {
			queryString = getQueryString(filterParams[0])
		}
		url := fmt.Sprintf("/v1/volumes%s", queryString)
		resp, err := c.get(url, nil, http.StatusOK, &vols)
		return vols, resp, err
	}
	var vol api.VolumeGetResp
	url := fmt.Sprintf("/v1/volumes/%s", volname)
	resp, err := c.get(url, nil, http.StatusOK, &vol)
	return []api.VolumeGetResp{vol}, resp, err
}

// BricksStatus returns the status of bricks that form a Gluster volume
func (c *Client) BricksStatus(volname string) (api.BricksStatusResp, *http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s/bricks", volname)
	var resp api.BricksStatusResp
	httpResp, err := c.get(url, nil, http.StatusOK, &resp)
	return resp, httpResp, err
}

// VolumeStatus returns the status of a Gluster volume
func (c *Client) VolumeStatus(volname string) (api.VolumeStatusResp, *http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s/status", volname)
	var volStatus api.VolumeStatusResp
	resp, err := c.get(url, nil, http.StatusOK, &volStatus)
	return volStatus, resp, err
}

// VolumeStart starts a Gluster Volume
func (c *Client) VolumeStart(volname string, force bool) (*http.Response, error) {
	req := api.VolumeStartReq{
		ForceStartBricks: force,
	}
	url := fmt.Sprintf("/v1/volumes/%s/start", volname)
	return c.post(url, req, http.StatusOK, nil)
}

// VolumeStop stops a Gluster Volume
func (c *Client) VolumeStop(volname string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s/stop", volname)
	return c.post(url, nil, http.StatusOK, nil)
}

// VolumeDelete deletes a Gluster Volume
func (c *Client) VolumeDelete(volname string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s", volname)
	return c.del(url, nil, http.StatusNoContent, nil)
}

// VolumeSet sets an option for a Gluster Volume
func (c *Client) VolumeSet(volname string, req api.VolOptionReq) (*http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s/options", volname)
	return c.post(url, req, http.StatusOK, nil)
}

// GlobalOptionSet sets cluster level options
func (c *Client) GlobalOptionSet(req api.GlobalOptionReq) (*http.Response, error) {
	url := fmt.Sprintf("/v1/cluster/options")
	return c.post(url, req, http.StatusOK, nil)
}

// VolumeGet gets volume options for a Gluster Volume
func (c *Client) VolumeGet(volname string, optname string) (api.VolumeOptionsGetResp, *http.Response, error) {
	if optname == "all" {
		var opts api.VolumeOptionsGetResp
		url := fmt.Sprintf("/v1/volumes/%s/options", volname)
		resp, err := c.get(url, nil, http.StatusOK, &opts)
		return opts, resp, err
	}
	var opt api.VolumeOptionGetResp
	url := fmt.Sprintf("/v1/volumes/%s/options/%s", volname, optname)
	resp, err := c.get(url, nil, http.StatusOK, &opt)
	return []api.VolumeOptionGetResp{opt}, resp, err
}

// VolumeExpand expands a Gluster Volume
func (c *Client) VolumeExpand(volname string, req api.VolExpandReq) (api.VolumeExpandResp, *http.Response, error) {
	var vol api.VolumeExpandResp
	url := fmt.Sprintf("/v1/volumes/%s/expand", volname)
	resp, err := c.post(url, req, http.StatusOK, &vol)
	return vol, resp, err
}

// VolumeStatedump takes statedump of various daemons
func (c *Client) VolumeStatedump(volname string, req api.VolStatedumpReq) (*http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s/statedump", volname)
	return c.post(url, req, http.StatusOK, nil)
}

// OptionGroupCreate creates a new option group
func (c *Client) OptionGroupCreate(req api.OptionGroupReq) (*http.Response, error) {
	return c.post("/v1/volumes/options-group", req, http.StatusOK, nil)
}

// OptionGroupList returns a list of all option groups
func (c *Client) OptionGroupList() (api.OptionGroupListResp, *http.Response, error) {
	var l api.OptionGroupListResp
	url := fmt.Sprintf("/v1/volumes/options-group")
	resp, err := c.get(url, nil, http.StatusOK, &l)
	return l, resp, err
}

// OptionGroupDelete deletes the specified option group
func (c *Client) OptionGroupDelete(group string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/options-group/%s", group)
	return c.del(url, nil, http.StatusNoContent, nil)
}

// EditVolume edits the specified keys in volinfo of a volume
func (c *Client) EditVolume(volname string, req api.VolEditReq) (api.VolumeEditResp, *http.Response, error) {
	var resp api.VolumeEditResp
	url := fmt.Sprintf("/v1/volumes/%s/edit", volname)
	httpResp, err := c.post(url, req, http.StatusOK, &resp)
	return resp, httpResp, err
}

// VolumeReset resets volume options to their default values
func (c *Client) VolumeReset(volname string, req api.VolOptionResetReq) (*http.Response, error) {
	url := fmt.Sprintf("/v1/volumes/%s/options", volname)
	return c.del(url, req, http.StatusOK, nil)
}
