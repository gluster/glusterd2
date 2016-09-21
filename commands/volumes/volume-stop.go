package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func stopBricks(c transaction.TxnCtx) error {
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

	for _, brick := range vol.Bricks {
		if uuid.Equal(brick.ID, gdctx.MyUUID) {
			c.Logger().WithFields(log.Fields{
				"volume": volname,
				"brick":  brick.Hostname + ":" + brick.Path,
			}).Info("would stop brick")
			//TODO: Stop running brick processes once the daemon management package is ready
		}
	}
	return nil
}

func registerVolStopStepFuncs() {
	transaction.RegisterStepFunc(stopBricks, "vol-stop.Commit")
}

func volumeStopHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume stop API")

	vol, e := volume.GetVolume(volname)
	if e != nil {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolNotFound.Error())
		return
	}
	if vol.Status == volume.VolStopped {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolAlreadyStopped.Error())
		return
	}

	// A simple one-step transaction to stop brick processes
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
			DoFunc: "vol-stop.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("volname", volname)

	_, e = txn.Do()
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e.Error(),
			"volume": volname,
		}).Error("failed to stop volume")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	vol.Status = volume.VolStopped

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	rest.SendHTTPResponse(w, http.StatusOK, vol)
}
