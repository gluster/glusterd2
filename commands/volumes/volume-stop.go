package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/daemon"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
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

	for _, b := range vol.Bricks {
		if uuid.Equal(b.NodeID, gdctx.MyUUID) {
			c.Logger().WithFields(log.Fields{
				"volume": volname,
				"brick":  b.Hostname + ":" + b.Path,
			}).Info("Stopping brick")

			brickDaemon, err := brick.NewGlusterfsd(b)
			if err != nil {
				return err
			}

			err = daemon.Stop(brickDaemon, false)
			if err != nil {
				return err
			}
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
	reqID := r.Header.Get("X-Request-ID")
	logger := log.WithField("reqid", reqID)

	vol, e := volume.GetVolume(volname)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}
	if vol.Status == volume.VolStopped {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolAlreadyStopped.Error())
		return
	}

	// A simple one-step transaction to stop brick processes
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
			DoFunc: "vol-stop.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("volname", volname)

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":  e.Error(),
			"volume": volname,
		}).Error("failed to stop volume")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	vol.Status = volume.VolStopped

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	restutils.SendHTTPResponse(w, http.StatusOK, vol)
}
