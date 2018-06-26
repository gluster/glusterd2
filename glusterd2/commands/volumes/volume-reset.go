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

func registerVolOptionResetStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-option.XlatorActionDoReset", xlatorActionDoReset},
		{"vol-option.XlatorActionUndoReset", xlatorActionUndoReset},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeResetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]

	var req api.VolOptionResetReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	//store volinfo to revert back changes in case of transaction failure
	oldvolinfo := volinfo
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
	//save volume information for transaction failure scenario
	if err := txn.Ctx.Set("oldvolinfo", oldvolinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	opt := make(map[string]string)
	for _, key := range req.Options {
		opt[key] = ""
	}
	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "vol-option.XlatorActionDoReset",
			UndoFunc: "vol-option.XlatorActionUndoReset",
			Nodes:    volinfo.Nodes(),
			Skip:     !isActionStepRequired(opt),
		},
		{
			DoFunc:   "vol-option.UpdateVolinfo",
			UndoFunc: "vol-option.UpdateVolinfo.Undo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  allNodes,
		},
	}
	if err := txn.Ctx.Set("req", &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volinfo.Options)
}
