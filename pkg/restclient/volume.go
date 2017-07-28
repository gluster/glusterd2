package restclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"
)

// VolumeCreate creates Gluster Volume
func (c *RESTClient) VolumeCreate(req api.VolCreateReq) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return httpRESTAction("POST", c.baseURL+"/v1/volumes", strings.NewReader(string(reqBody)), 201)
}

// VolumeStart starts a Gluster Volume
func (c *RESTClient) VolumeStart(volname string) error {
	url := fmt.Sprintf(c.baseURL+"/v1/volumes/%s/start", volname)
	return httpRESTAction("POST", url, nil, 200)
}

// VolumeStop stops a Gluster Volume
func (c *RESTClient) VolumeStop(volname string) error {
	url := fmt.Sprintf(c.baseURL+"/v1/volumes/%s/stop", volname)
	return httpRESTAction("POST", url, nil, 200)
}

// VolumeDelete deletes a Gluster Volume
func (c *RESTClient) VolumeDelete(volname string) error {
	url := fmt.Sprintf(c.baseURL+"/v1/volumes/%s", volname)
	return httpRESTAction("DELETE", url, nil, 200)
}
