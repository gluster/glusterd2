package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

func volumeInfoHandler(w http.ResponseWriter, r *http.Request) {

	v, err := volume.GetVolume(mux.Vars(r)["volname"])
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
	}

	resp := createVolumeGetResp(v)
	restutils.SendHTTPResponse(w, http.StatusOK, resp)
}

func createVolumeGetResp(v *volume.Volinfo) *api.VolumeGetResp {
	return (*api.VolumeGetResp)(createVolumeInfoResp(v))
}
