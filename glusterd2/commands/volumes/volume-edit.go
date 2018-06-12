package volumecommands

import (
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

func volumeEditHandler(w http.ResponseWriter, r *http.Request) {

	p := mux.Vars(r)
	volname := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req api.VolEditReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	//Lock on Volume Name
	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	defer txn.Done()

	//validate volume name
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	reqMetadataSize := req.MetadataSize()
	if reqMetadataSize > maxMetadataSizeLimit {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrMetadataSizeOutOfBounds)
		return
	}
	for key, value := range req.Metadata {
		if strings.HasPrefix(key, "_") {
			logger.WithField("key", key).Error(errors.ErrRestrictedKeyFound)
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrRestrictedKeyFound)
			return
		}
		if req.DeleteMetadata {
			delete(volinfo.Metadata, key)
		} else {
			volinfo.Metadata[key] = value
		}
	}

	if volinfo.MetadataSize() > maxMetadataSizeLimit {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrMetadataSizeOutOfBounds)
		return
	}
	if err := volume.AddOrUpdateVolumeFunc(volinfo); err != nil {
		logger.WithError(err).WithField(
			"volume", volinfo.Name).Debug("failed to store volume info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to store volume info")
		return
	}
	resp := createEditVolumeResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createEditVolumeResp(v *volume.Volinfo) *api.VolumeEditResp {
	return (*api.VolumeEditResp)(volume.CreateVolumeInfoResp(v))
}
