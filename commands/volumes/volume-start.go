package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/daemon"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func startBricks(c transaction.TxnCtx) error {
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

	for _, b := range vol.Bricks {
		if uuid.Equal(b.ID, gdctx.MyUUID) {
			c.Logger().WithFields(log.Fields{
				"volume": volname,
				"brick":  b.Hostname + ":" + b.Path,
			}).Info("Starting brick")

			brickDaemon, err := brick.NewDaemon(vol.Name, b)
			if err != nil {
				return err
			}

			err = daemon.Start(brickDaemon, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func undoStartBricks(c transaction.TxnCtx) error {
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

	for _, b := range vol.Bricks {
		if uuid.Equal(b.ID, gdctx.MyUUID) {
			c.Logger().WithFields(log.Fields{
				"volume": volname,
				"brick":  b.Hostname + ":" + b.Path,
			}).Info("volume start failed, stopping bricks")
			//TODO: Stop started brick processes once the daemon management package is ready

			brickDaemon, err := brick.NewDaemon(vol.Name, b)
			if err != nil {
				return err
			}

			err = daemon.Stop(brickDaemon, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func registerVolStartStepFuncs() {
	transaction.RegisterStepFunc(startBricks, "vol-start.Commit")
	transaction.RegisterStepFunc(undoStartBricks, "vol-start.Undo")
}

func volumeStartHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume start API")

	vol, e := volume.GetVolume(volname)
	if e != nil {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolNotFound.Error())
		return
	}
	if vol.Status == volume.VolStarted {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolAlreadyStarted.Error())
		return
	}

	// A simple one-step transaction to start the brick processes
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
			DoFunc:   "vol-start.Commit",
			UndoFunc: "vol-start.Undo",
			Nodes:    txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("volname", volname)

	_, e = txn.Do()
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e.Error(),
			"volume": volname,
		}).Error("failed to start volume")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	vol.Status = volume.VolStarted

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	rest.SendHTTPResponse(w, http.StatusOK, vol)
}
