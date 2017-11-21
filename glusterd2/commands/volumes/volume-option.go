package volumecommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/pborman/uuid"

	"github.com/gorilla/mux"
)

func registerVolOptionStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-option.UpdateVolinfo", storeVolume},
		{"vol-option.RegenerateVolfiles", generateBrickVolfiles},
		{"vol-option.NotifyVolfileChange", notifyVolfileChange},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeOptionsHandler(w http.ResponseWriter, r *http.Request) {

	p := mux.Vars(r)
	volname := p["volname"]
	reqID, logger := restutils.GetReqIDandLogger(r)

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	var req api.VolOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	if err := areOptionNamesValid(req.Options); err != nil {
		logger.WithField("option", err.Error()).Error("invalid option specified")
		restutils.SendHTTPError(w, http.StatusBadRequest, fmt.Sprintf("invalid option specified: %s", err.Error()), api.ErrCodeDefault)
		return
	}

	lock, unlock, err := transaction.CreateLockSteps(volinfo.Name)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()

	// thes txn framework checks if these nodes are online before txn starts
	txn.Nodes = volinfo.Nodes()

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.RegenerateVolfiles",
			// BUG: Shouldn't be on all nodes ideally. Currently we
			// can't know if it's a brick option or client option.
			// If it's a brick option, the nodes list here should
			// should be only volinfo.Nodes().
			Nodes: allNodes,
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  allNodes,
		},
		unlock,
	}

	for k, v := range req.Options {
		// TODO: Normalize <graph>.<xlator>.<option> and just
		// <xlator>.<option> to avoid ambiguity and duplication.
		// For example, currently both the following representations
		// will be stored in volinfo:
		// {"afr.eager-lock":"on","gfproxy.afr.eager-lock":"on"}
		volinfo.Options[k] = v
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if _, err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, volinfo.Options)
}
