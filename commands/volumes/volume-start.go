package volumecommands

import (
	"net/http"
	"os/exec"
	"syscall"
	"time"

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

// BrickStartMaxRetries represents maximum no. of attempts that will be made
// to start brick processes in case of port clashes.
const BrickStartMaxRetries = 3

// Until https://review.gluster.org/#/c/16200/ gets into a release.
// And this is fully safe too as no other well-known errno exists after 132
const anotherEADDRINUSE = syscall.Errno(0x9E) // 158

func errorContainsErrno(err error, errno syscall.Errno) bool {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == int(errno) {
				return true
			}
		}
	}
	return false
}

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

			for i := 0; i < BrickStartMaxRetries; i++ {
				err = daemon.Start(brickDaemon, true)
				if err != nil {
					if errorContainsErrno(err, syscall.EADDRINUSE) || errorContainsErrno(err, anotherEADDRINUSE) {
						// Retry iff brick failed to start because of port being in use.
						c.Logger().Info("Brick port unavailable. Retrying...")
						// Allow the previous instance to cleanup and exit
						time.Sleep(1 * time.Second)
					} else {
						return err
					}
				} else {
					break
				}
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
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolNotFound.Error())
		return
	}
	if vol.Status == volume.VolStarted {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolAlreadyStarted.Error())
		return
	}

	// A simple one-step transaction to start the brick processes
	txn := transaction.NewTxn()
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
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	vol.Status = volume.VolStarted

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	restutils.SendHTTPResponse(w, http.StatusOK, vol)
}
