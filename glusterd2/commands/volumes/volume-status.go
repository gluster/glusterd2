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
	volname := mux.Vars(r)["volname"]

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted)
		return
	}

	s, err := volume.UsageInfo(volinfo.Name)
	if err != nil {
		logger.WithError(err).WithField("volume", volinfo.Name).Error("Failed to get volume size info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get Volume size info")
		return

	}
	size := createSizeInfo(s)

	resp := createVolumeStatusResp(volinfo, &size)
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
