package bitrot

import (
	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	log "github.com/sirupsen/logrus"
)

// IsBitrotAffectedNode returns true if there are local bricks of volume on which bitrot is enabled
func IsBitrotAffectedNode() bool {
	volumes, e := volume.GetVolumes()
	if e != nil {
		log.WithError(e).Error("Failed to get volumes")
		return true
	}
	for _, v := range volumes {
		val, exists := v.Options[volume.VkeyFeaturesBitrot]
		if exists && val == "off" {
			continue
		} else if v.State != volume.VolStarted {
			continue
		} else {
			bricks := v.GetLocalBricks()
			if len(bricks) > 0 {
				return true
			}
			continue
		}
	}
	return false
}

// ManageBitd manages the bitrot daemon bitd. It stops or stops/starts the daemon based on disable or enable respectively.
func ManageBitd(bitrotDaemon *Bitd) error {
	var err error
	AffectedNode := IsBitrotAffectedNode()

	// Should the bitd and scrubber untouched on this node
	if !AffectedNode {
		// This condition is for disable
		// TODO: Need to ignore errors where process is already running.
		daemon.Stop(bitrotDaemon, true)
	} else {
		//TODO: Handle ENOENT of pidFile
		err = daemon.Stop(bitrotDaemon, true)
		err = daemon.Start(bitrotDaemon, true)
		if err != nil {
			return err
		}
	}
	return err
}

// ManageScrubd manages the scrubber daemon. It stops or stops/starts the daemon based on disable or enable respectively.
func ManageScrubd() error {
	AffectedNode := IsBitrotAffectedNode()
	scrubDaemon, err := newScrubd()
	if err != nil {
		return err
	}

	if !AffectedNode {
		// This condition is for disable
		// TODO: Need to ignore errors where process is already running.
		daemon.Stop(scrubDaemon, true)
	} else {
		//TODO: Handle ENOENT of pidFile
		daemon.Stop(scrubDaemon, true)
		err = daemon.Start(scrubDaemon, true)
		if err != nil {
			return err
		}
	}
	return err
}

func txnBitrotEnableDisable(c transaction.TxnCtx) error {
	bitrotDaemon, err := newBitd()
	if err != nil {
		return err
	}

	// Manange bitd and scrub daemons
	err = ManageBitd(bitrotDaemon)
	if err != nil {
		goto error
	}

	err = ManageScrubd()
	if err != nil {
		goto error
	}

	return nil
error:
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
	req.Input, err = dict.Serialize(reqDict)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to serialize dict for scrub-value")
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
