package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func volumeResetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotFound)
		return
	}

	var req api.VolOptionResetReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed)
		return
	}

	// Delete the option after checking for volopt flags
	opReset := false
	for _, k := range req.Options {
		// Check if the key is set or not
		if _, ok := volinfo.Options[k]; ok {
			op, err := xlator.FindOption(k)
			// If key exists, check for NEVER_RESET and FORCE flags
			if err != nil {
				restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
				return
			}
			if op.IsNeverReset() {
				errMsg := "Reserved option, can't be reset"
				restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
				return
			}
			if op.IsForceRequired() {
				if req.Force {
					delete(volinfo.Options, k)
					opReset = true
				} else {
					errMsg := "Option needs force flag to be set"
					restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
					return
				}
			}
			delete(volinfo.Options, k)
			opReset = true
		} else {
			errMsg := "Option trying to reset is not set or invalid option"
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
			return
		}
	}
	// Check if an option was reset, else return.
	if !opReset {
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, volinfo.Options)
		return
	}

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  allNodes,
		},
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volinfo.Options)
}
