package restclient

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
)

// SnapshotCreate creates Gluster Snapshot
func (c *Client) SnapshotCreate(req api.SnapCreateReq) (api.SnapCreateResp, error) {
	var snap api.SnapCreateResp
	err := c.post("/v1/snapshot", req, http.StatusCreated, &snap)
	return snap, err
}

//SnapshotActivate activate a Gluster snapshot
func (c *Client) SnapshotActivate(req api.SnapActivateReq, snapname string) error {
	url := fmt.Sprintf("/v1/snapshot/%s/activate", snapname)
	return c.post(url, req, http.StatusOK, nil)
}

//SnapshotDeactivate deactivate a Gluster snapshot
func (c *Client) SnapshotDeactivate(snapname string) error {
	url := fmt.Sprintf("/v1/snapshot/%s/deactivate", snapname)
	return c.post(url, nil, http.StatusOK, nil)
}

// SnapshotList returns list of all snapshots or all snapshots of a volume
func (c *Client) SnapshotList(volname string) (api.SnapListResp, error) {
	var snaps api.SnapListResp
	var url string
	if volname == "" {
		url = fmt.Sprintf("/v1/snapshots")
	} else {
		url = fmt.Sprintf("/v1/snapshots?volume=%s", volname)
	}
	err := c.get(url, nil, http.StatusOK, &snaps)
	return snaps, err
}

// SnapshotInfo returns information about a snapshot
func (c *Client) SnapshotInfo(snapname string) (api.SnapGetResp, error) {
	var snap api.SnapGetResp
	var url string
	url = fmt.Sprintf("/v1/snapshot/%s", snapname)
	err := c.get(url, nil, http.StatusOK, &snap)
	return snap, err
}

// SnapshotDelete will delete Gluster Snapshot and respective lv
func (c *Client) SnapshotDelete(snapname string) error {
	url := fmt.Sprintf("/v1/snapshot/%s", snapname)
	err := c.del(url, nil, http.StatusNoContent, nil)
	return err
}

// SnapshotStatus will list the detailed status of snapshot bricks
func (c *Client) SnapshotStatus(snapname string) (api.SnapStatusResp, error) {
	var resp api.SnapStatusResp

	url := fmt.Sprintf("/v1/snapshot/%s/status", snapname)
	err := c.get(url, nil, http.StatusOK, &resp)
	return resp, err
}

//SnapshotRestore will restore the volume to given snapshot
func (c *Client) SnapshotRestore(snapname string) (api.VolumeGetResp, error) {
	var resp api.VolumeGetResp
	url := fmt.Sprintf("/v1/snapshot/%s/restore", snapname)
	err := c.post(url, nil, http.StatusOK, &resp)
	return resp, err
}

// SnapshotClone creates a writable Gluster Snapshot, it will be similar to a volume
func (c *Client) SnapshotClone(snapname string, req api.SnapCloneReq) (api.VolumeCreateResp, error) {
	var vol api.VolumeCreateResp
	url := fmt.Sprintf("/v1/snapshot/%s/clone", snapname)
	err := c.post(url, req, http.StatusCreated, &vol)
	return vol, err
}
