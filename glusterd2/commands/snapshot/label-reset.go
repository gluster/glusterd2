package snapshotcommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot/label"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"

	"github.com/pborman/uuid"
)

func updateResetLabel(labelInfo *label.Info, req *api.LabelResetReq) (*label.Info, error) {

	for _, v := range req.Configurations {
		switch label.Options(v) {
		case label.SnapMaxHardLimitKey:
			labelInfo.SnapMaxHardLimit = label.DefaultLabel.SnapMaxHardLimit
		case label.SnapMaxSoftLimitKey:
			labelInfo.SnapMaxSoftLimit = label.DefaultLabel.SnapMaxSoftLimit
		case label.ActivateOnCreateKey:
			labelInfo.ActivateOnCreate = label.DefaultLabel.ActivateOnCreate
		case label.AutoDeleteKey:
			labelInfo.AutoDelete = label.DefaultLabel.AutoDelete
		default:
			return labelInfo, fmt.Errorf("%s is not a comptable ioption", v)

		}

	}
	return labelInfo, nil
}

func registerLabelConfigResetStepFuncs() {
	transaction.RegisterStepFunc(storeLabel, "label-config.Store")
}

func labelConfigResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	var req api.LabelResetReq

	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	labelname := mux.Vars(r)["labelname"]

	txn, err := transaction.NewTxnWithLocks(ctx, labelname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	labelInfo, err := label.GetLabel(labelname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	labelInfo, err = updateResetLabel(labelInfo, &req)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := validateLabel(labelInfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "label-config.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err = txn.Ctx.Set("label", &labelInfo); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("label config transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	labelInfo, err = label.GetLabel(labelname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Ctx.Logger().WithField("LabelName", labelname).Info("label modfied")

	resp := createLabelConfigResp(labelInfo)
	restutils.SetLocationHeader(r, w, labelInfo.Name)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)
}
