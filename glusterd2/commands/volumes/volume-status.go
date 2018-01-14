package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

func volumeStatusHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	v, err := volume.GetVolume(mux.Vars(r)["volname"])
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	s, err := volumeUsage(v.Name)
	if err != nil {
		logger.WithError(err).WithField("volume", v.Name).Error("Failed to get volume size info")
	}

	resp := createVolumeStatusResp(v, s)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeStatusResp(v *volume.Volinfo, s *api.SizeInfo) *api.VolumeStatusResp {
	resp := &api.VolumeStatusResp{
		Info: *(createVolumeInfoResp(v)),
	}

	if s != nil {
		// if the mount succeeded, then the volume is online
		resp.Online = true
		resp.Size = *s
	}

	return resp
}
