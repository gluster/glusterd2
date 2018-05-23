package rebalance

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"

	log "github.com/sirupsen/logrus"
)

type actionType uint16

const (
	actionStart actionType = iota
	actionStop
)

const (
	rebalStatusTxnKey string = "rebalstatus"
)

func txnRebalanceStart(c transaction.TxnCtx) error {
	var rinfo rebalanceapi.RebalInfo
	err := c.Get("rinfo", &rinfo)
	if err != nil {
		return err
	}

	rebalanceProcess, err := NewRebalanceProcess(rinfo)
	if err != nil {
		return err
	}

	err = daemon.Start(rebalanceProcess, true)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", rinfo.Volname).Error("Starting rebalance process failed")
		return err
	}

	return nil
}

func txnRebalanceStop(c transaction.TxnCtx) error {
	var rebalinfo rebalanceapi.RebalInfo
	err := c.Get("rinfo", &rebalinfo)
	if err != nil {
		return err
	}

	var volname string
	var xlatorName string
	var command string

	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	//TODO: Check rebalinfo status and reply if already finished.

	rebalanceProcess, err := NewRebalanceProcess(rebalinfo)
	if err != nil {
		return err
	}

	client, err := daemon.GetRPCClient(rebalanceProcess)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to connect to the rebalance process")
		return err
	}

	reqDict := make(map[string]string)

	xlatorName = fmt.Sprintf("%s-distribute", volname)

	command = fmt.Sprintf("%d", uint64(rebalanceapi.CmdStop))
	reqDict["rebalance-command"] = command

	log.Info("Stopping rebalance : command = ", command)

	req := &brick.GfBrickOpReq{
		Name: xlatorName,
		Op:   int(brick.OpBrickXlatorDefrag),
	}

	req.Input, err = dict.Serialize(reqDict)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to serialize dict for rebalance stop")
	}

	var rsp brick.GfBrickOpRsp
	err = client.Call("Brick.OpBrickXlatorDefrag", req, &rsp)
	if err != nil || rsp.OpRet != 0 {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to send rebalance stop RPC")
		return err
	}

	// TODO : Send a response
	// Unserialize the resp dict for rebalance stop
	//   rspDict, err := dict.Unserialize(rsp.Output)
	//    if err != nil {
	//             log.WithError(err).Error("dict unserialize failed")
	//              return err
	//       }

	return nil
}

func txnRebalanceStatus(c transaction.TxnCtx) error {

	var rebalinfo rebalanceapi.RebalInfo
	var rebalNodeStatus rebalanceapi.RebalNodeStatus

	err := c.Get("rinfo", &rebalinfo)
	if err != nil {
		return err
	}

	var volname string
	var xlatorName string
	var command string

	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	// What is the expected behaviour if the process does not exist (rebalance has completed)?
	// Will it restart the process?

	rebalanceProcess, err := NewRebalanceProcess(rebalinfo)
	if err != nil {

		// TODO: Send the stored Info?
		return nil
	}

	client, err := daemon.GetRPCClient(rebalanceProcess)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to connect to the rebalance process")

		/* Get status from store
		rebal, err := GetRebalanceInfo(volname)
		if err != nil {
			return err
		}

		rebalNodeStatus = rebal.RebalStats
		*/
		return err
	}

	//Send the status request to the rebalance process

	reqDict := make(map[string]string)

	//Set the name of the dht xlator
	xlatorName = fmt.Sprintf("%s-distribute", volname)

	command = fmt.Sprintf("%d", uint64(rebalanceapi.CmdStatus))
	reqDict["rebalance-command"] = command

	req := &brick.GfBrickOpReq{
		Name: xlatorName,
		Op:   int(brick.OpBrickXlatorDefrag),
	}
	req.Input, err = dict.Serialize(reqDict)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to serialize dict for rebalance status")
		return err
	}

	var rsp brick.GfBrickOpRsp
	err = client.Call("Brick.OpBrickXlatorDefrag", req, &rsp)
	if err != nil || rsp.OpRet != 0 {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to send rebalance status RPC")
		return err
	}

	// Unserialize the resp dict for rebalance status
	rspDict, err := dict.Unserialize(rsp.Output)
	if err != nil {
		log.WithError(err).Error("dict unserialize failed")
		return err
	}

	rebalNodeStatus.NodeID = gdctx.MyUUID
	rebalNodeStatus.Status = rspDict["status"]
	rebalNodeStatus.RebalancedFiles = rspDict["files"]
	rebalNodeStatus.RebalancedSize = rspDict["size"]
	rebalNodeStatus.LookedupFiles = rspDict["lookups"]
	rebalNodeStatus.SkippedFiles = rspDict["skipped"]
	rebalNodeStatus.RebalanceFailures = rspDict["failures"]
	rebalNodeStatus.ElapsedTime = rspDict["run-time"]

	c.SetNodeResult(gdctx.MyUUID, rebalStatusTxnKey, rebalNodeStatus)
	return nil

}

func txnRebalanceStoreDetails(c transaction.TxnCtx) error {
	var rebalinfo rebalanceapi.RebalInfo

	err := c.Get("rinfo", &rebalinfo)
	if err != nil {
		return err
	}

	err = StoreRebalanceInfo(&rebalinfo)
	if err != nil {
		log.WithError(err).Error("Couldn't add rebalance info to store")
		return err
	}

	return nil
}
