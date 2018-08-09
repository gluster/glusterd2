package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// SnapshotCreate creates Gluster Snapshot
func (c *Client) SnapshotCreate(req api.SnapCreateReq) (api.SnapCreateResp, *http.Response, error) {
	var snap api.SnapCreateResp
	resp, err := c.post("/v1/snapshot", req, http.StatusCreated, &snap)
	return snap, resp, err
}

//SnapshotActivate activate a Gluster snapshot
func (c *Client) SnapshotActivate(req api.SnapActivateReq, snapname string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/snapshot/%s/activate", snapname)
	return c.post(url, req, http.StatusOK, nil)
}

//SnapshotDeactivate deactivate a Gluster snapshot
func (c *Client) SnapshotDeactivate(snapname string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/snapshot/%s/deactivate", snapname)
	return c.post(url, nil, http.StatusOK, nil)
}

// SnapshotList returns list of all snapshots or all snapshots of a volume
func (c *Client) SnapshotList(volname string) (api.SnapListResp, *http.Response, error) {
	var snaps api.SnapListResp
	var url string
	if volname == "" {
		url = fmt.Sprintf("/v1/snapshots")
	} else {
		url = fmt.Sprintf("/v1/snapshots?volume=%s", volname)
	}
	resp, err := c.get(url, nil, http.StatusOK, &snaps)
	return snaps, resp, err
}

// SnapshotInfo returns information about a snapshot
func (c *Client) SnapshotInfo(snapname string) (api.SnapGetResp, *http.Response, error) {
	var snap api.SnapGetResp
	var url string
	url = fmt.Sprintf("/v1/snapshot/%s", snapname)
	resp, err := c.get(url, nil, http.StatusOK, &snap)
	return snap, resp, err
}

// SnapshotDelete will delete Gluster Snapshot and respective lv
func (c *Client) SnapshotDelete(snapname string) (*http.Response, error) {
	url := fmt.Sprintf("/v1/snapshot/%s", snapname)
	return c.del(url, nil, http.StatusNoContent, nil)
}

// SnapshotStatus will list the detailed status of snapshot bricks
func (c *Client) SnapshotStatus(snapname string) (api.SnapStatusResp, *http.Response, error) {
	var resp api.SnapStatusResp

	url := fmt.Sprintf("/v1/snapshot/%s/status", snapname)
	httpResp, err := c.get(url, nil, http.StatusOK, &resp)
	return resp, httpResp, err
}

//SnapshotRestore will restore the volume to given snapshot
func (c *Client) SnapshotRestore(snapname string) (api.VolumeGetResp, *http.Response, error) {
	var resp api.VolumeGetResp
	url := fmt.Sprintf("/v1/snapshot/%s/restore", snapname)
	httpResp, err := c.post(url, nil, http.StatusOK, &resp)
	return resp, httpResp, err
}

// SnapshotClone creates a writable Gluster Snapshot, it will be similar to a volume
func (c *Client) SnapshotClone(snapname string, req api.SnapCloneReq) (api.VolumeCreateResp, *http.Response, error) {
	var vol api.VolumeCreateResp
	url := fmt.Sprintf("/v1/snapshot/%s/clone", snapname)
	resp, err := c.post(url, req, http.StatusCreated, &vol)
	return vol, resp, err
}
