package restclient

import (
	"fmt"
	"net/http"

	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
)

// GeorepCreate establishes Geo-replication session
func (c *Client) GeorepCreate(mastervolid string, slavevolid string, req georepapi.GeorepCreateReq) (georepapi.GeorepSession, *http.Response, error) {
	var session georepapi.GeorepSession
	url := fmt.Sprintf("/v1/geo-replication/%s/%s", mastervolid, slavevolid)
	resp, err := c.post(url, req, http.StatusOK, &session)
	return session, resp, err
}

// GeorepStart starts Geo-replication session
func (c *Client) GeorepStart(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, *http.Response, error) {
	var session georepapi.GeorepSession
	opts := georepapi.GeorepCommandsReq{Force: force}
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/start", mastervolid, slavevolid)
	resp, err := c.post(url, &opts, http.StatusOK, &session)
	return session, resp, err
}

// GeorepPause pauses Geo-replication session
func (c *Client) GeorepPause(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, *http.Response, error) {
	var session georepapi.GeorepSession
	opts := georepapi.GeorepCommandsReq{Force: force}
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/pause", mastervolid, slavevolid)
	resp, err := c.post(url, &opts, http.StatusOK, &session)
	return session, resp, err
}

// GeorepResume resumes Geo-replication session
func (c *Client) GeorepResume(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, *http.Response, error) {
	var session georepapi.GeorepSession
	opts := georepapi.GeorepCommandsReq{Force: force}
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/resume", mastervolid, slavevolid)
	resp, err := c.post(url, &opts, http.StatusOK, &session)
	return session, resp, err
}

// GeorepStop stops Geo-replication session
func (c *Client) GeorepStop(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, *http.Response, error) {
	var session georepapi.GeorepSession
	opts := georepapi.GeorepCommandsReq{Force: force}
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/stop", mastervolid, slavevolid)
	resp, err := c.post(url, &opts, http.StatusOK, &session)
	return session, resp, err
}

// GeorepDelete deletes Geo-replication session
func (c *Client) GeorepDelete(mastervolid string, slavevolid string, force bool) (*http.Response, error) {
	opts := georepapi.GeorepCommandsReq{Force: force}
	url := fmt.Sprintf("/v1/geo-replication/%s/%s", mastervolid, slavevolid)
	return c.del(url, &opts, http.StatusNoContent, nil)

}

// GeorepStatus gets status of Geo-replication sessions
func (c *Client) GeorepStatus(mastervolid string, slavevolid string) ([]georepapi.GeorepSession, *http.Response, error) {
	var (
		err      error
		sessions []georepapi.GeorepSession
		httpResp *http.Response
	)

	url := "/v1/geo-replication"
	allSessions := false

	if mastervolid != "" && slavevolid != "" {
		allSessions = true
		url = fmt.Sprintf("%s/%s/%s", url, mastervolid, slavevolid)
	}

	if !allSessions {
		httpResp, err = c.get(url, nil, http.StatusOK, &sessions)
	} else {
		var session georepapi.GeorepSession
		httpResp, err = c.get(url, nil, http.StatusOK, &session)
		if err == nil {
			sessions = []georepapi.GeorepSession{session}
		}
	}
	return sessions, httpResp, err
}

// GeorepSSHKeysGenerate generates SSH keys in all Volume nodes
func (c *Client) GeorepSSHKeysGenerate(volname string) ([]georepapi.GeorepSSHPublicKey, *http.Response, error) {
	url := "/v1/ssh-key/" + volname + "/generate"
	var sshkeys []georepapi.GeorepSSHPublicKey
	resp, err := c.post(url, nil, http.StatusOK, &sshkeys)
	return sshkeys, resp, err
}

// GeorepSSHKeys gets SSH keys from all Volume nodes
func (c *Client) GeorepSSHKeys(volname string) ([]georepapi.GeorepSSHPublicKey, *http.Response, error) {
	url := "/v1/ssh-key/" + volname
	var sshkeys []georepapi.GeorepSSHPublicKey
	resp, err := c.get(url, nil, http.StatusOK, &sshkeys)
	return sshkeys, resp, err
}

// GeorepSSHKeysPush pushes SSH public keys to all Volume nodes
func (c *Client) GeorepSSHKeysPush(volname string, sshkeys []georepapi.GeorepSSHPublicKey) (*http.Response, error) {
	url := "/v1/ssh-key/" + volname + "/push"
	return c.post(url, sshkeys, http.StatusOK, nil)
}

// GeorepGet gets Geo-replication options
func (c *Client) GeorepGet(mastervolid string, slavevolid string) ([]georepapi.GeorepOption, *http.Response, error) {
	var options []georepapi.GeorepOption
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/config", mastervolid, slavevolid)
	resp, err := c.get(url, nil, http.StatusOK, &options)
	return options, resp, err
}

// GeorepSet sets Geo-replication options
func (c *Client) GeorepSet(mastervolid string, slavevolid string, keyvals map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/config", mastervolid, slavevolid)
	return c.post(url, &keyvals, http.StatusOK, nil)
}

// GeorepReset resets Geo-replication options
func (c *Client) GeorepReset(mastervolid string, slavevolid string, keys []string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/geo-replication/%s/%s/config", mastervolid, slavevolid)
	return c.del(url, &keys, http.StatusOK, nil)
}
