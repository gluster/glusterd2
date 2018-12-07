package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func prepareBricks(c transaction.TxnCtx) error {
	var req api.BrickReq
	if err := c.Get("newBrick", &req); err != nil {
		return err
	}
	err := PrepareBrick(req, c)
	return err
}

func replaceVolinfo(c transaction.TxnCtx) error {
	var newBrick api.BrickReq
	var err error
	if err = c.Get("newBrick", &newBrick); err != nil {
		return err
	}
	var srcBrickInfo brick.Brickinfo
	if err = c.Get("srcBrickInfo", &srcBrickInfo); err != nil {
		return err
	}
	var subVolIndex int
	if err = c.Get("subVolIndex", &subVolIndex); err != nil {
		return err
	}
	var brickIndex int
	if err = c.Get("brickIndex", &brickIndex); err != nil {
		return err
	}
	var volInfo volume.Volinfo
	if err = c.Get("volinfo", &volInfo); err != nil {
		return err
	}

	// Replace brick details in original VolInfo
	newBricks := []api.BrickReq{newBrick}
	newBrickInfos, err := volume.NewBrickEntries(newBricks, srcBrickInfo.VolumeName, srcBrickInfo.VolfileID, srcBrickInfo.VolumeID, srcBrickInfo.PType)
	if err != nil {
		return err
	}
	newBrickInfo := newBrickInfos[0]

	// Retaining the brick position same as old brick
	volInfo.Subvols[subVolIndex].Bricks[brickIndex] = newBrickInfo

	// Setting bricks Info in transaction context
	if err = c.Set("bricks", newBrickInfos); err != nil {
		return err
	}

	// Setting checks in transaction context
	//TODO: Ask for flags and force
	checks := brick.PrepareChecks(true, make(map[string]bool))
	err = c.Set("brick-checks", checks)
	if err != nil {
		return err
	}

	if err = c.Set("bricks", volInfo.GetBricks()); err != nil {
		return err
	}

	allBricks, err := volume.GetAllBricksInCluster()
	if err != nil {
		return err
	}

	// Used by other peers to check if proposed bricks are already in use.
	// This check is however still prone to races. See issue #314
	if err = c.Set("all-bricks-in-cluster", allBricks); err != nil {
		return err
	}

	// Setting volume Info in transaction context
	err = c.Set("volinfo", volInfo)
	return err
}

func startBrick(c transaction.TxnCtx) error {

	var newBrickInfo []brick.Brickinfo
	if err := c.Get("bricks", newBrickInfo); err != nil {
		return err
	}

	// Starting new brick
	err := newBrickInfo[0].StartBrick(c.Logger())
	return err
}
