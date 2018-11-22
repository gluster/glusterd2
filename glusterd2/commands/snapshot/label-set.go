package snapshotcommands

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot/label"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"

	"github.com/pborman/uuid"
)

func updateSetLabel(labelInfo *label.Info, req *api.LabelSetReq) (*label.Info, error) {

	for k, v := range req.Configurations {
		switch label.Options(k) {
		case label.SnapMaxHardLimitKey:
			value, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return labelInfo, fmt.Errorf("%s is not a comptable value for option %s", v, k)
			}
			labelInfo.SnapMaxHardLimit = value
		case label.SnapMaxSoftLimitKey:
			value, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return labelInfo, fmt.Errorf("%s is not a comptable value for option %s", v, k)
			}
			labelInfo.SnapMaxSoftLimit = value
		case label.ActivateOnCreateKey:
			value, err := strconv.ParseBool(v)
			if err != nil {
				return labelInfo, fmt.Errorf("%s is not a comptable value for option %s", v, k)
			}
			labelInfo.ActivateOnCreate = value
		case label.AutoDeleteKey:
			value, err := strconv.ParseBool(v)
			if err != nil {
				return labelInfo, fmt.Errorf("%s is not a comptable value for option %s", v, k)
			}
			labelInfo.AutoDelete = value
		default:
			return labelInfo, fmt.Errorf("%s is not a comptable option", k)

		}

	}
	return labelInfo, nil
}

func registerLabelConfigSetStepFuncs() {
	transaction.RegisterStepFunc(storeLabel, "label-config.Store")
}

func labelConfigSetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	var req api.LabelSetReq

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

	if labelname == (label.DefaultLabel).Name {
		errMsg := "Default label cannot be edited."
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	labelInfo, err := label.GetLabel(labelname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	labelInfo, err = updateSetLabel(labelInfo, &req)
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

	if err = txn.Ctx.Set("label", labelInfo); err != nil {
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
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createLabelConfigResp(info *label.Info) *api.LabelConfigResp {
	return (*api.LabelConfigResp)(label.CreateLabelInfoResp(info))
}
