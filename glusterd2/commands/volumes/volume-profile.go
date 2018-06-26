package volumecommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/gorilla/mux"
)

func registerVolProfileStepFuncs() {
	transaction.RegisterStepFunc(txnVolumeProfile, "volume.Profile")
}

func volumeProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]
	option := mux.Vars(r)["option"]

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Volume must be in stopped state before deleting.")
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "volume.Profile",
			Nodes:  volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("option", option); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to profile volume failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, nil)
}

func txnVolumeProfile(c transaction.TxnCtx) error {
	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	var option string
	if err := c.Get("option", &option); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(b)
		if err != nil {
			return err
		}

		c.Logger().WithField(
			"volume", volinfo.Name).Info("Starting volume profile operation")

		client, err := daemon.GetRPCClient(brickDaemon)
		if err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to connect to brick, sending SIGTERM")
			return err
		}
		reqDict := make(map[string]string)
		if option == "start" {
			reqDict["peek"] = "0"
			reqDict["op"] = "1"
			reqDict["info-op"] = "0"
		} else if option == "info" {
			reqDict["peek"] = "0"
			reqDict["op"] = "3"
			reqDict["info-op"] = "1"
		} else if option == "stop" {
			reqDict["peek"] = "0"
			reqDict["op"] = "2"
			reqDict["info-op"] = "0"
		}
		reqDict["volname"] = volinfo.Name
		reqDict["vol-id"] = volinfo.ID.String()
		req := &brick.GfBrickOpReq{
			Name: b.Path,
			Op:   int(brick.OpBrickXlatorInfo),
		}
		fmt.Println(req)
		fmt.Println(reqDict)
		req.Input, err = dict.Serialize(reqDict)
		var rsp brick.GfBrickOpRsp
		err = client.Call("Brick.OpBrickXlatorInfo", req, &rsp)
		if err != nil || rsp.OpRet != 0 {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to send volume profile RPC")
			return err
		}
		fmt.Println(rsp.Output)
	}
	return nil
}
