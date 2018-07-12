package glustershd

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
)

func getHxlChildrenCount(volinfo *volume.Volinfo) (int, string) {
	if volinfo.Type == volume.Replicate || volinfo.Type == volume.DistReplicate {
		return volinfo.Subvols[0].ReplicaCount, "replicate"
	}
	return volinfo.Subvols[0].DisperseCount, "disperse"
}

func addHxlatorToDict(reqDict map[string]string, volinfo *volume.Volinfo, index int, count int, xlType string) map[string]string {
	key := fmt.Sprintf("xl-%d", count)
	xName := fmt.Sprintf("%s-%s-%d", volinfo.Name, xlType, index)
	reqDict[key] = xName
	reqDict[xName] = strconv.Itoa(index)
	return reqDict
}

func selectHxlatorsWithBricks(volinfo *volume.Volinfo, healType int) map[string]string {
	index := 1
	hxlatorCount := 0
	add := false
	reqDict := make(map[string]string)
	reqDict["heal-op"] = strconv.Itoa(healType)
	reqDict["xl-op"] = reqDict["heal-op"]
	reqDict["volname"] = volinfo.Name
	reqDict["sync-mgmt-operation"] = strconv.Itoa(20)
	reqDict["vol-id"] = volinfo.ID.String()
	hxlChildren, xlType := getHxlChildrenCount(volinfo)
	volBricks := volinfo.GetBricks()
	for brick := range volBricks {
		hostKey := fmt.Sprintf("%d-hostname", index-1)
		brickPathKey := fmt.Sprintf("%d-path", index-1)
		reqDict[hostKey] = volBricks[brick].Hostname
		reqDict[brickPathKey] = volBricks[brick].Path
		if bytes.Equal(volBricks[brick].PeerID, gdctx.MyUUID) {
			add = true
		}
		if index%hxlChildren == 0 {
			if add {
				reqDict = addHxlatorToDict(reqDict, volinfo, (index-1)/hxlChildren, hxlatorCount, xlType)
				hxlatorCount++
			}
			add = false
		}
		index++
	}
	reqDict["count"] = strconv.Itoa(hxlatorCount)
	return reqDict
}

func txnSelfHeal(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	var healType int
	if err := c.Get("healType", &healType); err != nil {
		return err
	}

	volname := volinfo.Name

	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}

	c.Logger().WithField("volume", volname).Info("Starting Heal")

	client, err := daemon.GetRPCClient(glustershDaemon)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to connect to glustershd")
		return err
	}

	reqDict := make(map[string]string)
	req := &brick.GfBrickOpReq{
		Name: "",
		Op:   int(brick.OpBrickXlatorOp),
	}
	reqDict = selectHxlatorsWithBricks(&volinfo, healType)
	req.Input, err = dict.Serialize(reqDict)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to serialize dict for index heal")
		return err
	}

	var rsp brick.GfBrickOpRsp
	err = client.Call("Brick.OpBrickXlatorOp", req, &rsp)
	if err != nil || rsp.OpRet != 0 {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to send index heal RPC")
		return err
	}

	return nil
}
