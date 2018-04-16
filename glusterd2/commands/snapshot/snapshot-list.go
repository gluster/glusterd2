package snapshotcommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gorilla/mux"
)

func snapshotListHandler(w http.ResponseWriter, r *http.Request) {

	volumeName := mux.Vars(r)["volumename"]
	snapName := make(map[string][]string)
	ctx := r.Context()

	if volumeName != "" {
		vol, e := volume.GetVolume(volumeName)
		if e != nil {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, e)
			return
		}
		snapName[volumeName] = vol.SnapList
	} else {

		snaps, err := snapshot.GetSnapshots()
		if err != nil {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		}
		for _, s := range snaps {
			snapName[s.ParentVolume] = append(snapName[s.ParentVolume], s.SnapVolinfo.Name)
		}
	}
	resp := createSnapshotListResp(snapName)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createSnapshotListResp(snaps map[string][]string) *api.SnapListResp {
	var resp api.SnapListResp
	var entry api.SnapList

	for key, s := range snaps {
		entry.ParentName = key
		entry.SnapName = s
		resp = append(resp, entry)
	}

	return &resp
}
