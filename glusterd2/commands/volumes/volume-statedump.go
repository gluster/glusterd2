package volumecommands

import (
	"errors"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/plugins/quota"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"golang.org/x/sys/unix"
)

func validateVolStatedumpReq(req *api.VolStatedumpReq) error {

	var tmp api.VolStatedumpReq // zero value of struct
	if *req == tmp {
		return errors.New("at least one of the statedump req options must be set")
	}

	if req.Client != tmp.Client {
		_, err := govalidator.ValidateStruct(req)
		if err != nil {
			return err
		}
	}

	return nil
}

func takeStatedump(c transaction.TxnCtx) error {

	var req api.VolStatedumpReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if req.Client.Host != "" && req.Client.Pid != 0 {
		sunrpc.ClientStatedump(volinfo.Name, req.Client.Host, req.Client.Pid, c.Logger())
	}

	if req.Bricks {
		for _, b := range volinfo.GetLocalBricks() {
			d, err := brick.NewGlusterfsd(b)
			if err != nil {
				return err
			}
			if err := daemon.Signal(d, unix.SIGUSR1, c.Logger()); err != nil {
				// only log, don't error out
				c.Logger().WithError(err).WithField(
					"daemon", d.ID()).Error("Failed to take statedump for daemon")
			}
		}
	}

	if req.Quota {
		d, err := quota.NewQuotad()
		if err != nil {
			return err
		}
		if err := daemon.Signal(d, unix.SIGUSR1, c.Logger()); err != nil {
			// only log, don't error out
			c.Logger().WithError(err).WithField(
				"daemon", d.ID()).Error("Failed to take statedump for daemon")
		}
	}

	return nil
}

func registerVolStatedumpFuncs() {
	transaction.RegisterStepFunc(takeStatedump, "vol-statedump.TakeStatedump")
}

func volumeStatedumpHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	var req api.VolStatedumpReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	if err := validateVolStatedumpReq(&req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrVolNotStarted)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-statedump.TakeStatedump",
			Nodes:  volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("req", &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to take statedump failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
