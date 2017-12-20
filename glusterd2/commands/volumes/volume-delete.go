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

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		return err
	}

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	if err := volgen.DeleteClientVolfile(volinfo); err != nil {
		// Log and continue, ignore the cleanup error
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Warn("deleteVolfiles: failed to delete client volfile")
	}

	for _, subvol := range volinfo.Subvols {
		for _, b := range subvol.Bricks {
			if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
				continue
			}

			if err := volgen.DeleteBrickVolfile(&b); err != nil {
				// Log and continue, ignore the volfile cleanup error
				c.Logger().WithError(err).WithField(
					"brick", b.Path).Warn("deleteVolfiles: failed to delete brick volfile")
			}
		}
	}

	return nil
}

func deleteVolume(c transaction.TxnCtx) error {

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		return err
	}

	return volume.DeleteVolume(volname)
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
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
		return
	}

	if vol.State == volume.VolStarted {
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
			Nodes:  vol.Nodes(),
		},
		{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	txn.Ctx.Set("volname", volname)
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

	events.Broadcast(newVolumeEvent(eventVolumeDeleted, vol))
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
