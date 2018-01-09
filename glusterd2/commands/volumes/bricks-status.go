package volumecommands

import (
	"net/http"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

const (
	brickStatusTxnKey string = "brickstatuses"
)

func checkBricksStatus(ctx transaction.TxnCtx) error {

	var volname string
	if err := ctx.Get("volname", &volname); err != nil {
		ctx.Logger().WithError(err).Error("Failed to get key from transaction context.")
		return err
	}

	vol, err := volume.GetVolume(volname)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get volume information from store.")
		return err
	}

	mtabEntries, err := getMounts()
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to read /etc/mtab file.")
		return err
	}

	var brickStatuses []*api.BrickStatus
	for _, binfo := range vol.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(binfo)
		if err != nil {
			return err
		}

		s := &api.BrickStatus{
			Info: createBrickInfo(&binfo),
		}

		if pidOnFile, err := daemon.ReadPidFromFile(brickDaemon.PidFile()); err == nil {
			if _, err := daemon.GetProcess(pidOnFile); err == nil {
				s.Online = true
				s.Pid = pidOnFile
				s.Port = pmap.RegistrySearch(binfo.Path, pmap.GfPmapPortBrickserver)
			}
		}

		var fstat syscall.Statfs_t
		if err := syscall.Statfs(binfo.Path, &fstat); err != nil {
			ctx.Logger().WithError(err).WithField("path",
				binfo.Path).Error("syscall.Statfs() failed")
		} else {
			s.Size = *(createSizeInfo(&fstat))
		}

		for _, m := range mtabEntries {
			if strings.HasPrefix(binfo.Path, m.mntDir) {
				s.MountOpts = m.mntOpts
				s.Device = m.fsName
				s.FS = m.mntType
			}
		}

		brickStatuses = append(brickStatuses, s)
	}

	// Store the results in transaction context. This will be consumed by
	// the node that initiated the transaction.
	ctx.SetNodeResult(gdctx.MyUUID, brickStatusTxnKey, brickStatuses)
	return nil
}

func registerBricksStatusStepFuncs() {
	transaction.RegisterStepFunc(checkBricksStatus, "bricks-status.Check")
}

func volumeBricksStatusHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "bricks-status.Check",
			Nodes:  vol.Nodes(),
		},
	}
	txn.Ctx.Set("volname", volname)

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("volume", volname).Error("Failed to get volume status")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	result, err := createBricksStatusResp(txn.Ctx, vol)
	if err != nil {
		errMsg := "Failed to aggregate brick status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("volumeStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, errMsg, api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, result)
}

func createBricksStatusResp(ctx transaction.TxnCtx, vol *volume.Volinfo) (*api.BricksStatusResp, error) {

	// bmap is a map of brick statuses keyed by brick ID
	bmap := make(map[string]*api.BrickStatus)
	for _, b := range vol.GetBricks() {
		bmap[b.ID.String()] = &api.BrickStatus{
			Info: createBrickInfo(&b),
		}
	}

	// Loop over each node that make up the volume and aggregate result
	// of brick status check from each.
	var resp api.BricksStatusResp
	for _, node := range vol.Nodes() {
		var tmp []api.BrickStatus
		err := ctx.GetNodeResult(node, brickStatusTxnKey, &tmp)
		if err != nil || len(tmp) == 0 {
			// skip if we do not have information
			continue
		}
		for _, b := range tmp {
			bmap[b.Info.ID.String()] = &b
		}
	}

	for _, v := range bmap {
		resp = append(resp, *v)
	}

	return &resp, nil
}
