package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

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
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("deleteVolfiles: failed to delete client volfile")
		return err
	}

	for _, b := range volinfo.Bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		if err := volgen.DeleteBrickVolfile(&b); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("deleteVolfiles: failed to delete brick volfile")
			return err
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
	p := mux.Vars(r)
	volname := p["volname"]
	reqID, logger := restutils.GetReqIDandLogger(r)

	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
		return
	}

	if vol.State == volume.VolStarted {
		restutils.SendHTTPError(w, http.StatusForbidden, "volume is not stopped", api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-delete.Commit",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	txn.Ctx.Set("volname", volname)
	if _, err = txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("failed to delete the volume")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, nil)
}
