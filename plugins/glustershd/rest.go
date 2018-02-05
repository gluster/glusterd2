package glustershd

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func isVolReplicate(vType volume.VolType) bool {
	if vType == volume.Replicate || vType == volume.Disperse || vType == volume.DistReplicate || vType == volume.DistDisperse {
		return true
	}

	return false
}

func glustershEnableHandler(w http.ResponseWriter, r *http.Request) {
	// Implement the help logic and send response back as below
	p := mux.Vars(r)
	volname := p["name"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	//validate volume name
	v, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}
	// Store initial volinfo before changing the HealFlag
	tmp := *v
	oldvolinfo := &tmp

	// validate volume type
	if !isVolReplicate(v.Type) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume Type not supported", api.ErrCodeDefault)
		return
	}

	if v.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume should be in started state.", api.ErrCodeDefault)
		return

	}

	// Transaction which starts self heal daemon on all nodes with atleast one brick.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	//Lock on Volume Name
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	v.HealEnabled = true

	txn.Nodes = v.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "vol-option.UpdateVolinfo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
			UndoFunc: "selfheald-undo",
		},
		{
			DoFunc: "selfheal-start",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	if err := txn.Ctx.Set("volinfo", v); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Ctx.Set("oldvolinfo", oldvolinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("failed to start self heal daemon")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func glustershDisableHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["name"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	//validate volume name
	v, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}
	// Store initial volinfo before changing the HealFlag
	tmp := *v
	oldvolinfo := &tmp

	// validate volume type
	if !isVolReplicate(v.Type) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume Type not supported", api.ErrCodeDefault)
		return
	}

	// Transaction which checks if all replicate volumes are stopped before
	// stopping the self-heal daemon.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	// Lock on volume name.
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	v.HealEnabled = false

	txn.Nodes = v.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "vol-option.UpdateVolinfo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
			UndoFunc: "selfheald-undo",
		},

		{
			DoFunc: "selfheal-stop",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	if err := txn.Ctx.Set("volinfo", v); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Ctx.Set("oldvolinfo", oldvolinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to stop self heal daemon")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
