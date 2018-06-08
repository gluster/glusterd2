package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
)

func volumeInfoHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	volname := mux.Vars(r)["volname"]
	v, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	resp := createVolumeGetResp(v)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeGetResp(v *volume.Volinfo) *api.VolumeGetResp {
	return (*api.VolumeGetResp)(volume.CreateVolumeInfoResp(v))
}
