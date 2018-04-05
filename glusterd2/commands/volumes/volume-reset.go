package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func registerVolResetStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-reset.Store", storeVolume},
		{"vol-reset.NotifyVolfileChange", notifyVolfileChange},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeResetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]
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

	txn := transaction.NewTxn(ctx)
	// Delete the option after checking for volopt flags
	success := false
	for _, k := range req.Options {
		if _, ok := volinfo.Options[k]; ok {
			if !volinfo.Options[k].VOLOPT_FLAG_NEVER_RESET {
				if !volinfo.Options[k].VOLOPT_FLAG_FORCE {
					delete(volinfo.Options, k)
					success = true
				} else if volinfo.Options[k].VOLOPT_FLAG_FORCE && req.Force {
					delete(volinfo.Options, k)
					success = true
				}
			}
		} else {
			logger.WithError(err).Error("Option trying to reset is not set or invalid option")
		}
	}
	// Check if an option was reset, else return.
	if !success {
		return
	}

	lock, unlock, err := transaction.CreateLockSteps(volinfo.Name)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-reset.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-reset.NotifyVolfileChange",
			Nodes:  allNodes,
		},
		unlock,
	}

	// Reset the Options with new values
	for key, value := range req.Options {
		volinfo.Options[key] = value
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
