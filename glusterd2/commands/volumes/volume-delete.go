package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deleteVolfiles(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := volgen.DeleteClientVolfile(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Warn("failed to delete client volfile")
	}

	for _, b := range volinfo.GetLocalBricks() {
		if err := volgen.DeleteBrickVolfile(&b); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Warn("failed to delete brick volfile")
		}
	}

	return nil
}

func deleteVolume(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	return volume.DeleteVolume(volinfo.Name)
}

func registerVolDeleteStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-delete.DeleteVolfiles", deleteVolfiles},
		{"vol-delete.Store", deleteVolume},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeDeleteHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

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

	if volinfo.State == volume.VolStarted {
		errMsg := "Volume must be in stopped state before deleting."
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-delete.DeleteVolfiles",
			Nodes:  volinfo.Nodes(),
		},
		{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to delete volume failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume deleted")
	events.Broadcast(volume.NewEvent(volume.EventVolumeDeleted, volinfo))

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
