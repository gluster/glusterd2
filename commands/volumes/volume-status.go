package volumecommands

import (
	goerrors "errors"
	"net/http"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/servers/rest"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

const (
	brickStatusTxnKey string = "brickstatuses"
)

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

	var brickStatuses []*brick.Brickstatus

	for _, binfo := range vol.Bricks {
		// Skip bricks that aren't on this node.
		// TODO: Rename Brickinfo field 'ID' to 'NodeUUID'
		if uuid.Equal(binfo.ID, gdctx.MyUUID) == false {
			continue
		}

		// TODO: Check actual brick status when we get them running.
		fakeStatus := &brick.Brickstatus{
			Hostname: binfo.Hostname,
			Path:     binfo.Path,
			ID:       binfo.ID,
			Online:   false,
			Pid:      1234,
		}
		brickStatuses = append(brickStatuses, fakeStatus)
	}

	// Store the results in transaction context. This will be consumed by
	// the node that initiated the transaction.
	ctx.SetNodeResult(gdctx.MyUUID, brickStatusTxnKey, brickStatuses)

	return nil
}

func registerVolStatusStepFuncs() {
	transaction.RegisterStepFunc(checkStatus, "vol-status.Check")
}

func aggregateVolumeStatus(ctx transaction.TxnCtx, nodes []uuid.UUID) (*volume.VolStatus, error) {
	var brickStatuses []brick.Brickstatus

	// Loop over each node on which txn was run.
	// Fetch brick statuses stored by each node in transaction context.
	for _, node := range nodes {
		var tmp []brick.Brickstatus
		err := ctx.GetNodeResult(node, brickStatusTxnKey, &tmp)
		if err != nil {
			return nil, goerrors.New("aggregateVolumeStatus: Could not fetch results from transaction context.")
		}
		brickStatuses = append(brickStatuses, tmp...)
	}
	v := &volume.VolStatus{Brickstatuses: brickStatuses}
	return v, nil
}

func volumeStatusHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	// Ensure that the volume exists.
	vol, err := volume.GetVolume(volname)
	if err != nil {
		rest.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
	}

	// A very simple free-form transaction to query each node for brick
	// status. Fetching volume status does not modify state/data on the
	// remote node. So there's no need for locks.
	txn := transaction.NewTxn()
	defer txn.Cleanup()
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		&transaction.Step{
			DoFunc: "vol-status.Check",
			Nodes:  txn.Nodes,
		},
	}

	// The remote nodes get args it needs from the transaction context.
	txn.Ctx.Set("volname", volname)

	// As all key-value pairs stored in transaction context ends up in etcd
	// store, using either the old txn context reference or the one
	// returned by txn.Do() is one and the same. The transaction context is
	// a way for the nodes store the results of the step runs.
	rtxn, err := txn.Do()
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err.Error(),
			"volume": volname,
		}).Error("volumeStatusHandler: Failed to get volume status.")
		rest.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Example of how an aggregate function will make sense from results of
	// run of a step on multiple nodes. The transaction context will have
	// results from each node, seggregated by the node's UUID.
	result, err := aggregateVolumeStatus(rtxn, txn.Nodes)
	if err != nil {
		errMsg := "Failed to aggregate brick status results from multiple nodes."
		log.WithField("error", err.Error()).Error("volumeStatusHandler:" + errMsg)
		rest.SendHTTPError(w, http.StatusInternalServerError, errMsg)
		return
	}

	// Send aggregated result back to the client.
	rest.SendHTTPResponse(w, http.StatusOK, result)
}
