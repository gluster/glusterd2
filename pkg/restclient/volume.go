package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// VolumeCreate creates Gluster Volume
func (c *Client) VolumeCreate(req api.VolCreateReq) (api.Volinfo, error) {
	var vol api.Volinfo
	err := c.post("/v1/volumes", req, http.StatusCreated, &vol)
	return vol, err
}

// Volumes returns list of all volumes
func (c *Client) Volumes() (api.VolList, error) {
	var vols api.VolList
	err := c.get("/v1/volumes", nil, http.StatusOK, &vols)
	return vols, err
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
func (c *Client) VolumeSet(volname string, options string) error {
	url := fmt.Sprintf("/v1/volumes/%s/options", volname)
	err := c.post(url, nil, http.StatusOK, nil)
	return err
}
