package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/volume"

	"github.com/gorilla/mux"
)

func volumeInfoHandler(w http.ResponseWriter, r *http.Request) {

	v, err := volume.GetVolume(mux.Vars(r)["volname"])
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
	}

	resp := createVolumeGetResp(v)
	restutils.SendHTTPResponse(w, http.StatusOK, resp)
}

func createVolumeGetResp(v *volume.Volinfo) *api.VolumeGetResp {
	return (*api.VolumeGetResp)(createVolumeInfoResp(v))
}
