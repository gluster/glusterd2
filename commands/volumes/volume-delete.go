package volumecommands

import (
	"errors"
	"net/http"

	gderrors "github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
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
	reqID, logger := restutils.GetReqIDandLogger(r)

	if !volume.Exists(volname) {
		restutils.SendHTTPError(w, http.StatusBadRequest, gderrors.ErrVolNotFound.Error())
		return
	}
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if vol.Status == volume.VolStarted {
		restutils.SendHTTPError(w, http.StatusForbidden, errors.New("volume is not stopped").Error())
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
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
		logger.WithFields(log.Fields{
			"error":  e.Error(),
			"volume": volname,
		}).Error("Failed to delete the volume")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	restutils.SendHTTPResponse(w, http.StatusOK, nil)
}
