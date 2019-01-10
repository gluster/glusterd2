package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/plugins/blockvolume/api"
)

// BlockVolumeCreate creates Gluster Block Volume
func (c *Client) BlockVolumeCreate(provider string, req api.BlockVolumeCreateRequest) (api.BlockVolumeCreateResp, error) {
	var vol api.BlockVolumeCreateResp
	err := c.post("v1/blockvolumes/"+provider, req, http.StatusCreated, &vol)
	return vol, err
}

// BlockVolumeList lists Gluster Block Volumes
func (c *Client) BlockVolumeList(provider string) (api.BlockVolumeListResp, error) {
	//TODO: Are filters required?
	var vols api.BlockVolumeListResp
	url := fmt.Sprintf("/v1/blockvolumes/%s", provider)
	err := c.get(url, nil, http.StatusOK, &vols)
	return vols, err
}

// BlockVolumeGet gets Gluster Block Volume info
func (c *Client) BlockVolumeGet(provider string, blockVolname string) (api.BlockVolumeGetResp, error) {
	var vol api.BlockVolumeGetResp
	url := fmt.Sprintf("/v1/blockvolumes/%s/%s", provider, blockVolname)
	err := c.get(url, nil, http.StatusOK, &vol)
	return vol, err
}

// BlockVolumeDelete deletes Gluster Block Volume
func (c *Client) BlockVolumeDelete(provider string, blockVolname string) error {
	url := fmt.Sprintf("/v1/blockvolumes/%s/%s", provider, blockVolname)
	return c.del(url, nil, http.StatusNoContent, nil)
}
