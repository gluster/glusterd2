package restclient

import (
	"fmt"
	"net/http"

	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
)

// GeorepCreate establishes Geo-replication session
func (c *Client) GeorepCreate(mastervolid string, slavevolid string, req georepapi.GeorepCreateReq) (georepapi.GeorepSession, error) {
	var session georepapi.GeorepSession
	url := fmt.Sprintf("/v1/geo-replication/%s/%s", mastervolid, slavevolid)
	err := c.post(url, req, http.StatusCreated, &session)
	return session, err
}
