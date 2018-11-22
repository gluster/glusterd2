package snapshotcommands

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot/label"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
)

const maxSnapCount = 256

//LoadDefaultLabel will store default label into etcd store. If already present then will not do anything
func LoadDefaultLabel() error {
	labelInfo := label.DefaultLabel

	if label.ExistsFunc(labelInfo.Name) {
		return nil
	}

	err := label.AddOrUpdateLabelFunc(&labelInfo)
	return err
}

func validateLabel(info *label.Info) error {

	if info.SnapMaxHardLimit > maxSnapCount {
		return fmt.Errorf("Snap-max-hard-limit count cannot exceed more than %d", maxSnapCount)
	}
	if info.SnapMaxSoftLimit > info.SnapMaxHardLimit {
		return errors.New("snap-soft-limit cannot exceed more than snap-max-hard-limit")
	}
	return nil
}

func newLabelInfo(req *api.LabelCreateReq) *label.Info {
	var labelInfo label.Info

	labelInfo.Name = req.Name
	labelInfo.SnapMaxHardLimit = req.SnapMaxHardLimit
	labelInfo.SnapMaxSoftLimit = req.SnapMaxSoftLimit
	labelInfo.ActivateOnCreate = req.ActivateOnCreate
	labelInfo.AutoDelete = req.AutoDelete
	labelInfo.Description = req.Description

	return &labelInfo
}

func storeLabel(c transaction.TxnCtx) error {

	var labelInfo label.Info

	if err := c.Get("label", &labelInfo); err != nil {
		return err
	}
	if err := label.AddOrUpdateLabelFunc(&labelInfo); err != nil {
		c.Logger().WithError(err).WithField(
			"label", labelInfo.Name).Debug("storeLabel: failed to store label info")
		return err
	}

	return nil
}

func registerLabelCreateStepFuncs() {
	transaction.RegisterStepFunc(storeLabel, "label-create.Store")
}

func labelCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	var req api.LabelCreateReq

	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}
	if label.ExistsFunc(req.Name) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrLabelExists)
		return
	}

	/*
		TODO : label name validation
	*/

	labelInfo := newLabelInfo(&req)
	if err := validateLabel(labelInfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, req.Name)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "label-create.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err = txn.Ctx.Set("label", &labelInfo); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("label create transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	labelInfo, err = label.GetLabel(req.Name)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Ctx.Logger().WithField("LabelName", req.Name).Info("new label created")

	resp := createLabelCreateResp(labelInfo)
	restutils.SetLocationHeader(r, w, labelInfo.Name)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)
}

func createLabelCreateResp(info *label.Info) *api.LabelCreateResp {
	return (*api.LabelCreateResp)(label.CreateLabelInfoResp(info))
}
