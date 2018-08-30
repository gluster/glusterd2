package snapshotcommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func snapshotListHandler(w http.ResponseWriter, r *http.Request) {

	snapName := make(map[string][]api.SnapInfo)
	ctx := r.Context()

	volumeName := r.URL.Query().Get("volume")

	if volumeName != "" {
		vol, err := volume.GetVolume(volumeName)
		if err != nil {
			status, err := restutils.ErrToStatusCode(err)
			restutils.SendHTTPError(ctx, w, status, err)
			return
		}
		for _, s := range vol.SnapList {
			snapInfo, err := snapshot.GetSnapshot(s)
			if err != nil {
				status, err := restutils.ErrToStatusCode(err)
				restutils.SendHTTPError(ctx, w, status, err)
				return
			}
			snapName[volumeName] = append(snapName[volumeName], *createSnapInfoResp(snapInfo))
		}

	} else {
		snaps, err := snapshot.GetSnapshots()
		if err != nil {
			status, err := restutils.ErrToStatusCode(err)
			restutils.SendHTTPError(ctx, w, status, err)
			return
		}
		for _, s := range snaps {
			snapName[s.ParentVolume] = append(snapName[s.ParentVolume], *createSnapInfoResp(s))
		}
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, createSnapshotListResp(snapName))
}

func createSnapshotListResp(volSnaps map[string][]api.SnapInfo) *api.SnapListResp {
	var resp api.SnapListResp
	resp = make(api.SnapListResp, 0)
	for vol, snapList := range volSnaps {
		var snap api.SnapList
		snap.ParentName = vol
		for _, s := range snapList {
			snap.SnapList = append(snap.SnapList, s)
		}
		resp = append(resp, snap)
	}
	return &resp
}
