package snapshotcommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot/label"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func registerLabelDeleteStepFuncs() {
	transaction.RegisterStepFunc(deleteLabel, "label-delete.Store")
}

func deleteLabel(c transaction.TxnCtx) error {

	var labelInfo label.Info
	if err := c.Get("labelinfo", &labelInfo); err != nil {
		return err
	}

	err := label.DeleteLabel(&labelInfo)
	return err
}

func labelDeleteHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	labelname := mux.Vars(r)["labelname"]
	labelInfo, err := label.GetLabel(labelname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, labelname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	if labelname == (label.DefaultLabel).Name {
		errMsg := "Default label cannot be deleted."
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	if len(labelInfo.SnapList) > 0 {
		errMsg := fmt.Sprintf("Cannot delete Label %s ,as it has %d snapshots tagged.", labelname, len(labelInfo.SnapList))
		restutils.SendHTTPError(ctx, w, http.StatusFailedDependency, errMsg)
		return
	}
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "label-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("labelinfo", labelInfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"label", labelname).Error("transaction to delete label failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("label-name", labelname).Info("label deleted")
	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}
