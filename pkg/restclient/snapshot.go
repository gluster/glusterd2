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
func (c *Client) SnapshotList(req api.SnapListReq) (api.SnapListResp, error) {
	var snaps api.SnapListResp
	url := fmt.Sprintf("/v1/snapshots")
	err := c.get(url, req, http.StatusOK, &snaps)
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
	err := c.del(url, nil, http.StatusOK, nil)
	return err
}
