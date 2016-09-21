package volumecommands

import (
	"errors"
	"net/http"

	gderrors "github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deleteVolfiles(c transaction.TxnCtx) error {
	var volname string

	if e := c.Get("volname", &volname); e != nil {
		c.Logger().WithFields(log.Fields{
			"error": e,
			"key":   "volname",
		}).Error("failed to get value for key from context")
		return e
	}

	vol, e := volume.GetVolume(volname)
	if e != nil {
		// this shouldn't happen
		c.Logger().WithFields(log.Fields{
			"error":   e,
			"volname": volname,
		}).Error("failed to get volinfo for volume")
		return e
	}

	return volgen.DeleteVolfile(vol)
}

func deleteVolume(c transaction.TxnCtx) error {
	var volname string

	if e := c.Get("volname", &volname); e != nil {
		c.Logger().WithFields(log.Fields{
			"error": e,
			"key":   "volname",
		}).Error("failed to get value for key from context")
		return e
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

	log.Info("In Volume delete API")

	if !volume.Exists(volname) {
		rest.SendHTTPError(w, http.StatusBadRequest, gderrors.ErrVolNotFound.Error())
		return
	}
	vol, err := volume.GetVolume(volname)
	if err != nil {
		// this shouldn't happen
		rest.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if vol.Status == volume.VolStarted {
		rest.SendHTTPError(w, http.StatusForbidden, errors.New("volume is not stopped").Error())
		return
	}

	// This is an example of a freeform transaction, that doesn't use the simple transaction framework

	txn := transaction.NewTxn()
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		&transaction.Step{
			DoFunc: "vol-delete.Commit",
			Nodes:  txn.Nodes,
		},
		&transaction.Step{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	txn.Ctx.Set("volname", volname)

	_, e := txn.Do()
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e.Error(),
			"volume": volname,
		}).Error("Failed to delete the volume")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	rest.SendHTTPResponse(w, http.StatusOK, nil)
}
