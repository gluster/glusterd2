package quota

import (
	"net/http"
	"os"
	"path"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

//QuotadStart is the function to start the quota daemon
func QuotadStart(c transaction.TxnCtx) error {
	quotadDaemon, err := NewQuotad()
	if err != nil {
		return err
	}
	// Create pidfile dir if not exists
	err = os.MkdirAll(path.Dir(quotadDaemon.pidfilepath), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}
	// Create logFiledir dir
	err = os.MkdirAll(path.Dir(quotadDaemon.logfilepath), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}
	err = daemon.Start(quotadDaemon, true, c.Logger())
	if err == errors.ErrProcessAlreadyRunning {
		c.Logger().WithError(err).Warn("Quota Daemon is already running.")
		return nil
	}
	return err
}

func quotaEnableHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volName := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	txn, err := transaction.NewTxnWithLocks(ctx, volName)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate volume existence
	vol, err := volume.GetVolume(volName)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	// Check if volume is started
	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted)
		return
	}

	// Check if quota is already enabled
	if isQuotaEnabled(vol) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrProcessAlreadyRunning)
		return
	}

	// Enable quota
	vol.Options[quotaEnabledKey] = "on"

	if err := txn.Ctx.Set("volinfo", vol); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "quota-enable.DaemonStart",
			Nodes:  vol.Nodes(),
		},
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("quota enable transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn.Ctx.Logger().WithField("volname", volName).Info("quota enabled")

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "quota enabled")
}

func quotaDisableHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "todo: quota disable")
}

func quotaListHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "todo: quota list")
}

func quotaLimitHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Todo: limitusage")
}

func quotaRemoveHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Todo: quota Remove")
}
