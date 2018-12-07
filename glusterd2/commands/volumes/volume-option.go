package volumecommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/options"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

func optionSetValidate(c transaction.TxnCtx) error {

	var req api.VolOptionReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	options, err := expandGroupOptions(req.Options)
	if err != nil {
		return err
	}

	// TODO: Validate op versions of the options. Either here or inside
	// validateOptions.

	if err := validateOptions(options, req.VolOptionFlags); err != nil {
		return fmt.Errorf("validation failed for volume option: %s", err.Error())
	}

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := validateXlatorOptions(options, &volinfo); err != nil {
		return fmt.Errorf("validation failed for volume option:: %s", err.Error())
	}

	for k, v := range options {
		// TODO: Normalize <graph>.<xlator>.<option> and just
		// <xlator>.<option> to avoid ambiguity and duplication.
		// For example, currently both the following representations
		// will be stored in volinfo:
		// {"afr.eager-lock":"on","gfproxy.afr.eager-lock":"on"}
		volinfo.Options[k] = v
	}

	err = c.Set("volinfo", volinfo)

	return err
}

type txnOpType uint8

const (
	txnDo txnOpType = iota
	txnUndo
)

func xlatorActionDoSet(c transaction.TxnCtx) error {
	return xlatorAction(c, txnDo, xlator.VolumeSet)
}

func xlatorActionUndoSet(c transaction.TxnCtx) error {
	return xlatorAction(c, txnUndo, xlator.VolumeSet)
}

func xlatorActionDoReset(c transaction.TxnCtx) error {
	return xlatorAction(c, txnDo, xlator.VolumeReset)
}

func xlatorActionUndoReset(c transaction.TxnCtx) error {
	return xlatorAction(c, txnUndo, xlator.VolumeReset)
}

func xlatorActionDoVolumeStart(c transaction.TxnCtx) error {
	return xlatorAction(c, txnDo, xlator.VolumeStart)
}

func xlatorActionUndoVolumeStart(c transaction.TxnCtx) error {
	return xlatorAction(c, txnUndo, xlator.VolumeStart)
}

func xlatorActionDoVolumeStop(c transaction.TxnCtx) error {
	return xlatorAction(c, txnDo, xlator.VolumeStop)
}

func xlatorActionUndoVolumeStop(c transaction.TxnCtx) error {
	return xlatorAction(c, txnUndo, xlator.VolumeStop)
}

// This function can be reused when volume reset operation is implemented.
// However, volume reset can be also be treated logically as volume set but
// with the value set to default value.
func xlatorAction(c transaction.TxnCtx, txnOp txnOpType, volOp xlator.VolumeOpType) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	reqOptions := make(map[string]string)
	var fn func(*volume.Volinfo, string, string, xlator.VolumeOpType, log.FieldLogger) error
	switch volOp {
	case xlator.VolumeSet:
		var req api.VolOptionReq
		if err := c.Get("req", &req); err != nil {
			return err
		}
		for key, value := range req.Options {
			reqOptions[key] = value
		}
	case xlator.VolumeReset:
		var req api.VolOptionResetReq
		if err := c.Get("req", &req); err != nil {
			return err
		}
		for _, key := range req.Options {
			op, err := xlator.FindOption(key)
			if err != nil {
				return err
			}
			reqOptions[key] = op.DefaultValue
		}

	case xlator.VolumeStart:
		fallthrough

	case xlator.VolumeStop:
		for _, actor := range xlator.GetOptActors() {
			if txnOp == txnDo {
				fn = actor.Do
			} else {
				fn = actor.Undo
			}
			if err := fn(&volinfo, "", "", volOp, c.Logger()); err != nil {
				return err
			}
		}
		return nil
	}
	//The code below applies only for xlator.VolumeSet and xlator.VolumeReset operations.
	for k, v := range reqOptions {
		_, xl, key := options.SplitKey(k)
		xltr, err := xlator.Find(xl)
		if err != nil {
			return err
		}
		if xltr.Actor != nil {
			if txnOp == txnDo {
				fn = xltr.Actor.Do
			} else {
				fn = xltr.Actor.Undo
			}
			if err := fn(&volinfo, key, v, volOp, c.Logger()); err != nil {
				return err
			}
		}
	}

	return nil
}

func registerVolOptionStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-option.Validate", optionSetValidate},
		{"vol-option.XlatorActionDoSet", xlatorActionDoSet},
		{"vol-option.XlatorActionUndoSet", xlatorActionUndoSet},
		{"vol-option.UpdateVolinfo", storeVolume},
		{"vol-option.UpdateVolinfo.Undo", undoStoreVolume},
		{"vol-option.NotifyVolfileChange", notifyVolfileChange},
		{"vol-option.GenerateBrickVolfiles", txnGenerateBrickVolfiles},
		{"vol-option.GenerateBrickvolfiles.Undo", txnDeleteBrickVolfiles},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeOptionsHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, span := trace.StartSpan(ctx, "/volumeOptionsHandler")
	defer span.End()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	var req api.VolOptionReq
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

	//save volume information for transaction failure scenario
	if err := txn.Ctx.Set("oldvolinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-option.Validate",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "vol-option.UpdateVolinfo",
			UndoFunc: "vol-option.UpdateVolinfo.Undo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "vol-option.XlatorActionDoSet",
			UndoFunc: "vol-option.XlatorActionUndoSet",
			Nodes:    volinfo.Nodes(),
			Skip:     !isActionStepRequired(req.Options, volinfo),
		},
		{
			DoFunc:   "vol-option.GenerateBrickVolfiles",
			UndoFunc: "vol-option.GenerateBrickvolfiles.Undo",
			Nodes:    volinfo.Nodes(),
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
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Add relevant attributes to the span
	var optionToSet string
	for option, value := range req.Options {
		optionToSet += option + "=" + value + ","
	}

	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("volName", volname),
		trace.StringAttribute("optionToSet", optionToSet),
	)

	if err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	volinfo, err = volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume options changed")

	resp := createVolumeOptionResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeOptionResp(v *volume.Volinfo) *api.VolumeOptionResp {
	return (*api.VolumeOptionResp)(volume.CreateVolumeInfoResp(v))
}
