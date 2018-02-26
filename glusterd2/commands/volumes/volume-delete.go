package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"

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
		{"vol-delete.Commit", deleteVolfiles},
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
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
		return
	}

	if volinfo.State == volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusForbidden, "volume is not stopped", api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-delete.Commit",
			Nodes:  volinfo.Nodes(),
		},
		{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	txn.Ctx.Set("volinfo", volinfo)
	if err = txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("failed to delete the volume")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	events.Broadcast(newVolumeEvent(eventVolumeDeleted, volinfo))
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
