package volumecommands

import (
	"net/http"

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
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	brickStatusTxnKey string = "brickstatuses"
)

type brickstatus struct {
	Info   brick.Brickinfo
	Online bool
	Pid    int
	Port   int
	// TODO: Add other fields like filesystem type, statvfs output etc.
}

func checkStatus(ctx transaction.TxnCtx) error {
	var volname string

	if err := ctx.Get("volname", &volname); err != nil {
		ctx.Logger().WithFields(log.Fields{
			"error": err,
			"key":   "volname",
		}).Error("checkStatus: Failed to get key from transaction context.")
		return err
	}

	vol, err := volume.GetVolume(volname)
	if err != nil {
		ctx.Logger().WithFields(log.Fields{
			"error": err,
			"key":   "volname",
		}).Error("checkStatus: Failed to get volume information from store.")
		return err
	}

	var brickStatuses []*brickstatus

	for _, binfo := range vol.Bricks {
		// Skip bricks that aren't on this node.
		if uuid.Equal(binfo.NodeID, gdctx.MyUUID) == false {
			continue
		}

		var port int

		brickDaemon, err := brick.NewGlusterfsd(binfo)
		if err != nil {
			return err
		}

		online := false

		pid, err := daemon.ReadPidFromFile(brickDaemon.PidFile())
		if err == nil {
			// Check if process is running
			_, err := daemon.GetProcess(pid)
			if err == nil {
				online = true
				port = pmap.RegistrySearch(binfo.Path, pmap.GfPmapPortBrickserver)
			}
		}

		brickStatus := &brickstatus{
			Info:   binfo,
			Online: online,
			Pid:    pid,
			Port:   port,
		}
		brickStatuses = append(brickStatuses, brickStatus)
	}

	// Store the results in transaction context. This will be consumed by
	// the node that initiated the transaction.
	ctx.SetNodeResult(gdctx.MyUUID, brickStatusTxnKey, brickStatuses)

	return nil
}

func registerVolStatusStepFuncs() {
	transaction.RegisterStepFunc(checkStatus, "vol-status.Check")
}

func volumeStatusHandler(w http.ResponseWriter, r *http.Request) {

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
	volNodes := vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-status.Check",
			Nodes:  volNodes,
		},
	}

	txn.Ctx.Set("volname", volname)

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":  err.Error(),
			"volume": volname,
		}).Error("volumeStatusHandler: Failed to get volume status.")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	result, err := createVolumeStatusResp(txn.Ctx, vol)
	if err != nil {
		errMsg := "Failed to aggregate brick status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("volumeStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, errMsg, api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, result)
}

func createVolumeStatusResp(ctx transaction.TxnCtx, vol *volume.Volinfo) (*api.VolumeStatusResp, error) {

	bmap := make(map[string]*api.BrickStatus)

	for _, b := range vol.Bricks {
		bmap[b.ID.String()] = &api.BrickStatus{
			Info: createBrickInfo(&b),
		}
	}

	// Loop over each node on which txn was run.
	// Fetch brick statuses stored by each node in transaction context.
	for _, node := range vol.Nodes() {
		var tmp []brickstatus
		err := ctx.GetNodeResult(node, brickStatusTxnKey, &tmp)
		if err != nil || len(tmp) == 0 {
			// skip if we do not have information
			continue
		}
		for _, b := range tmp {
			entry := bmap[b.Info.ID.String()]
			entry.Online = b.Online
			entry.Port = b.Port
			entry.Pid = b.Pid
		}
	}

	var resp api.VolumeStatusResp
	for _, v := range bmap {
		resp.Bricks = append(resp.Bricks, *v)
	}

	return &resp, nil
}
