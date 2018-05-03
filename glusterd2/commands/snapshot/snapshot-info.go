package snapshotcommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
)

func snapshotInfoHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	volname := mux.Vars(r)["snapname"]
	snap, err := snapshot.GetSnapshot(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		return
	}

	resp := createSnapGetResp(snap)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

// createSnapCreateResp functions create resnse for rest utils
func createSnapGetResp(snap *snapshot.Snapinfo) *api.SnapGetResp {
	return (*api.SnapGetResp)(createSnapInfoResp(snap))
}
