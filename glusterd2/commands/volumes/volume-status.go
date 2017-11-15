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
	p := mux.Vars(r)
	volname := p["volname"]
	reqID, logger := restutils.GetReqIDandLogger(r)

	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-status.Check",
			Nodes:  txn.Nodes,
		},
	}

	txn.Ctx.Set("volname", volname)

	rtxn, err := txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":  err.Error(),
			"volume": volname,
		}).Error("volumeStatusHandler: Failed to get volume status.")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := createVolumeStatusResp(rtxn, txn.Nodes)
	if err != nil {
		errMsg := "Failed to aggregate brick status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("volumeStatusHandler:" + errMsg)
		restutils.SendHTTPError(w, http.StatusInternalServerError, errMsg)
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, result)
}

func createVolumeStatusResp(ctx transaction.TxnCtx, nodes []uuid.UUID) (*api.VolumeStatusResp, error) {

	// Loop over each node on which txn was run.
	// Fetch brick statuses stored by each node in transaction context.
	var brickStatuses []brickstatus
	for _, node := range nodes {
		var tmp []brickstatus
		err := ctx.GetNodeResult(node, brickStatusTxnKey, &tmp)
		if err != nil {
			return nil, err
		}
		brickStatuses = append(brickStatuses, tmp...)
	}

	var resp api.VolumeStatusResp
	for _, b := range brickStatuses {
		resp.Bricks = append(resp.Bricks, api.BrickStatus{
			Info:   createBrickInfo(&b.Info),
			Online: b.Online,
			Pid:    b.Pid,
			Port:   b.Port,
		})
	}

	return &resp, nil
}
