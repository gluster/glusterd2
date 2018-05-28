package glustershd

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
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

	//validate volume name
	v, err := volume.GetVolume(volname)
	if err != nil {
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	// Store initial volinfo before changing the HealFlag
	tmp := *v
	oldvolinfo := &tmp

	// validate volume type
	if !isVolReplicate(v.Type) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume Type not supported")
		return
	}

	if v.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume should be in started state.")
		return

	}

	v.HealEnabled = true

	txn.Nodes = v.Nodes()
	txn.Steps = []*transaction.Step{
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
	}

	if err := txn.Ctx.Set("volinfo", v); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("oldvolinfo", oldvolinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("failed to start self heal daemon")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func glustershDisableHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["name"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

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

	//validate volume name
	v, err := volume.GetVolume(volname)
	if err != nil {
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	// Store initial volinfo before changing the HealFlag
	tmp := *v
	oldvolinfo := &tmp

	// validate volume type
	if !isVolReplicate(v.Type) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume Type not supported")
		return
	}

	v.HealEnabled = false

	txn.Nodes = v.Nodes()
	txn.Steps = []*transaction.Step{
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
	}

	if err := txn.Ctx.Set("volinfo", v); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("oldvolinfo", oldvolinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to stop self heal daemon")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
