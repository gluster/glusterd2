package restclient

import (
	"encoding/json"
	"fmt"
	"strings"
)

type volCreateReq struct {
	Name      string   `json:"name"`
	Transport string   `json:"transport,omitempty"`
	Replica   int      `json:"replica,omitempty"`
	Bricks    []string `json:"bricks"`
	Force     bool     `json:"force,omitempty"`
}

func (c *RESTClient) VolumeCreate(volname string, bricks []string, replica int, force bool) error {
	createReq := volCreateReq{
		Name:    volname,
		Replica: replica,
		Bricks:  bricks,
		Force:   force,
	}
	reqBody, err := json.Marshal(createReq)
	if err != nil {
		return err
	}
	return httpRESTAction("POST", c.baseURL+"/v1/volumes", strings.NewReader(string(reqBody)), 201)
}

func (c *RESTClient) VolumeStart(volname string) error {
	url := fmt.Sprintf(c.baseURL+"/v1/volumes/%s/start", volname)
	return httpRESTAction("POST", url, nil, 200)
}

func (c *RESTClient) VolumeStop(volname string) error {
	url := fmt.Sprintf(c.baseURL+"/v1/volumes/%s/stop", volname)
	return httpRESTAction("POST", url, nil, 200)
}

func (c *RESTClient) VolumeDelete(volname string) error {
	url := fmt.Sprintf(c.baseURL+"/v1/volumes/%s", volname)
	return httpRESTAction("DELETE", url, nil, 200)
}
