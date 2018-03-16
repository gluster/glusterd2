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

	if v.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted.Error(), api.ErrCodeDefault)
		return
	}

	s, err := volume.UsageInfo(v.Name)
	if err != nil {
		logger.WithError(err).WithField("volume", v.Name).Error("Failed to get volume size info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get Volume size info", api.ErrCodeDefault)
		return

	}
	size := createSizeInfo(s)

	resp := createVolumeStatusResp(v, &size)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeStatusResp(v *volume.Volinfo, s *api.SizeInfo) *api.VolumeStatusResp {
	resp := &api.VolumeStatusResp{
		Info: *(volume.CreateVolumeInfoResp(v)),
	}

	if s != nil {
		// if the mount succeeded, then the volume is online
		resp.Online = true
		resp.Size = *s
	}

	return resp
}
