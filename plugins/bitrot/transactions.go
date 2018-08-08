package bitrot

import (
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
	bitrotapi "github.com/gluster/glusterd2/plugins/bitrot/api"
	log "github.com/sirupsen/logrus"
)

const (
	scrubStatusTxnKey string = "scrubstatus"
)

// IsBitrotAffectedNode returns true if there are local bricks of volume on which bitrot is enabled
func IsBitrotAffectedNode() bool {
	volumes, e := volume.GetVolumes()
	if e != nil {
		log.WithError(e).Error("Failed to get volumes")
		return false
	}
	for _, v := range volumes {
		val, exists := v.Options[keyFeaturesBitrot]
		if exists && val == "off" || v.State != volume.VolStarted {
			continue
		}
		if len(v.GetLocalBricks()) > 0 {
			return true
		}
	}
	return false
}

// ManageBitd manages the bitrot daemon bitd. It stops or stops/starts the daemon based on disable or enable respectively.
func ManageBitd(bitrotDaemon *Bitd, logger log.FieldLogger) error {

	// Should the bitd and scrubber untouched on this node
	if !IsBitrotAffectedNode() {
		// This condition is for disable
		// TODO: Need to ignore errors where process is already running.
		if err := daemon.Stop(bitrotDaemon, true, logger); err != errors.ErrPidFileNotFound {
			return err
		}
		return nil
	}
	//TODO: Handle ENOENT of pidFile
	if err := daemon.Stop(bitrotDaemon, true, logger); err !=
		errors.ErrPidFileNotFound {
		return err
	}
	if err := daemon.Start(bitrotDaemon, true, logger); err != errors.ErrProcessAlreadyRunning {
		return err
	}
	return nil
}

// ManageScrubd manages the scrubber daemon. It stops or stops/starts the daemon based on disable or enable respectively.
func ManageScrubd(logger log.FieldLogger) error {
	scrubDaemon, err := newScrubd()
	if err != nil {
		return err
	}

	if !IsBitrotAffectedNode() {
		// This condition is for disable
		// TODO: Need to ignore errors where process is already running.
		if err = daemon.Stop(scrubDaemon, true, logger); err != errors.ErrPidFileNotFound {
			return err
		}
		return nil
	}
	//TODO: Handle ENOENT of pidFile
	if err = daemon.Stop(scrubDaemon, true, logger); err != errors.ErrPidFileNotFound {
		return err
	}
	if err = daemon.Start(scrubDaemon, true, logger); err != errors.ErrProcessAlreadyRunning {
		return err
	}

	return nil
}

func txnBitrotEnableDisable(c transaction.TxnCtx) error {
	bitrotDaemon, err := newBitd()
	if err != nil {
		return err
	}

	// Manange bitd and scrub daemons
	if err = ManageBitd(bitrotDaemon, c.Logger()); err != nil {
		goto Err
	}

	if err = ManageScrubd(c.Logger()); err != nil {
		goto Err
	}

	return nil
Err:
	//TODO: Handle failure of scrubd. bitd should be stopped. Should it be handled in txn undo func
	return err
}

func txnBitrotScrubOndemand(c transaction.TxnCtx) error {

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	scrubDaemon, err := newScrubd()
	if err != nil {
		return err
	}

	c.Logger().WithFields(log.Fields{"volume": volname}).Info("Starting scrubber")

	client, err := daemon.GetRPCClient(scrubDaemon)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to connect to scrubd")
		return err
	}

	reqDict := make(map[string]string)
	reqDict["scrub-value"] = "ondemand"
	req := &brick.GfBrickOpReq{
		Name: volname,
		Op:   int(brick.OpNodeBitrot),
	}

	if req.Input, err = dict.Serialize(reqDict); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to serialize dict for scrub-value")
		return err
	}

	var rsp brick.GfBrickOpRsp
	err = client.Call("Brick.OpNodeBitrot", req, &rsp)
	if err != nil || rsp.OpRet != 0 {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to send scrubondemand RPC")
		return err
	}

	return nil
}

func txnBitrotScrubStatus(c transaction.TxnCtx) error {

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	scrubDaemon, err := newScrubd()
	if err != nil {
		return err
	}

	client, err := daemon.GetRPCClient(scrubDaemon)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to connect to scrubd")
		return err
	}

	reqDict := make(map[string]string)
	reqDict["scrub-value"] = "status"
	req := &brick.GfBrickOpReq{
		Name: volname,
		Op:   int(brick.OpNodeBitrot),
	}
	if req.Input, err = dict.Serialize(reqDict); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to serialize dict for scrub-value")
		return err
	}
	var rsp brick.GfBrickOpRsp
	err = client.Call("Brick.OpNodeBitrot", req, &rsp)
	if err != nil || rsp.OpRet != 0 {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to send scrubstatus RPC")
		return err
	}

	// Unserialize the resp dict for scrub status
	rspDict, err := dict.Unserialize(rsp.Output)
	if err != nil {
		log.WithError(err).Error("dict unserialize failed")
		return err
	}

	var scrubNodeInfo bitrotapi.ScrubNodeInfo
	scrubNodeInfo.Node = gdctx.MyUUID.String()
	scrubNodeInfo.ScrubRunning = rspDict["scrub-running"]
	scrubNodeInfo.NumScrubbedFiles = rspDict["scrubbed-files"]
	scrubNodeInfo.NumSkippedFiles = rspDict["unsigned-files"]
	scrubNodeInfo.LastScrubCompletedTime = rspDict["last-scrub-time"]
	scrubNodeInfo.LastScrubDuration = rspDict["scrub-duration"]
	scrubNodeInfo.ErrorCount = rspDict["total-count"]

	// Fill CorruptedObjects
	count, err := strconv.Atoi(scrubNodeInfo.ErrorCount)
	if err != nil {
		log.WithError(err).Error("strconv of total-count failed")
		return err
	}
	for i := 0; i < count; i++ {
		countStr := strconv.Itoa(i)
		scrubNodeInfo.CorruptedObjects = append(scrubNodeInfo.CorruptedObjects, rspDict["quarantine-"+countStr])
	}

	// Store the results in transaction context. This will be consumed by
	// the node that initiated the transaction.
	c.SetNodeResult(gdctx.MyUUID, scrubStatusTxnKey, scrubNodeInfo)
	return nil
}
