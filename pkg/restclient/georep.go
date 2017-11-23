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

// GeorepStart starts Geo-replication session
func (c *Client) GeorepStart(mastervolid string, slavevolid string) (georepapi.GeorepSession, error) {
	var session georepapi.GeorepSession
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/start", mastervolid, slavevolid)
	err := c.post(url, nil, http.StatusOK, &session)
	return session, err
}

// GeorepPause pauses Geo-replication session
func (c *Client) GeorepPause(mastervolid string, slavevolid string) (georepapi.GeorepSession, error) {
	var session georepapi.GeorepSession
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/pause", mastervolid, slavevolid)
	err := c.post(url, nil, http.StatusOK, &session)
	return session, err
}

// GeorepResume resumes Geo-replication session
func (c *Client) GeorepResume(mastervolid string, slavevolid string) (georepapi.GeorepSession, error) {
	var session georepapi.GeorepSession
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/resume", mastervolid, slavevolid)
	err := c.post(url, nil, http.StatusOK, &session)
	return session, err
}

// GeorepStop stops Geo-replication session
func (c *Client) GeorepStop(mastervolid string, slavevolid string) (georepapi.GeorepSession, error) {
	var session georepapi.GeorepSession
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/stop", mastervolid, slavevolid)
	err := c.post(url, nil, http.StatusOK, &session)
	return session, err
}

// GeorepDelete deletes Geo-replication session
func (c *Client) GeorepDelete(mastervolid string, slavevolid string) error {
	url := fmt.Sprintf("/v1/geo-replication/%s/%s", mastervolid, slavevolid)
	err := c.del(url, nil, http.StatusOK, nil)
	return err
}
