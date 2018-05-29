package volumecommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
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

	if err := validateOptions(options, req.Advanced, req.Experimental, req.Deprecated); err != nil {
		return fmt.Errorf("Validation failed for volume option: %s", err.Error())
	}

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := validateXlatorOptions(req.Options, &volinfo); err != nil {
		return fmt.Errorf("Validation failed for volume option:: %s", err.Error())
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
	return xlatorAction(c, txnDo)
}

func xlatorActionUndoSet(c transaction.TxnCtx) error {
	return xlatorAction(c, txnUndo)
}

// This function can be reused when volume reset operation is implemented.
// However, volume reset can be also be treated logically as volume set but
// with the value set to default value.
func xlatorAction(c transaction.TxnCtx, txnOp txnOpType) error {

	var req api.VolOptionReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	var fn func(*volume.Volinfo, string, string) error
	for k, v := range req.Options {
		_, xl, key, err := options.SplitKey(k)
		if err != nil {
			return err
		}
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
			if err := fn(&volinfo, key, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func isActionStepRequired(req *api.VolOptionReq) bool {

	for k := range req.Options {
		_, xl, _, err := options.SplitKey(k)
		if err != nil {
			continue
		}
		if xltr, err := xlator.Find(xl); err == nil && xltr.Actor != nil {
			return true
		}
	}

	return false
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
		{"vol-option.NotifyVolfileChange", notifyVolfileChange},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeOptionsHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	var req api.VolOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

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
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
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
			DoFunc:   "vol-option.XlatorActionDoSet",
			UndoFunc: "vol-option.XlatorActionUndoSet",
			Nodes:    volinfo.Nodes(),
			Skip:     isActionStepRequired(&req),
		},
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
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

	if err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	volinfo, err = volume.GetVolume(volname)
	if err != nil {
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume options changed")

	resp := createVolumeOptionResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeOptionResp(v *volume.Volinfo) *api.VolumeOptionResp {
	return (*api.VolumeOptionResp)(volume.CreateVolumeInfoResp(v))
}
