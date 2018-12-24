package volumecommands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/brickmux"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func expandValidatePrepare(c transaction.TxnCtx) error {

	var req api.VolExpandReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		return err
	}

	//As of now volume expand doesn't support auto provisioned bricks
	provisionType := brick.ManuallyProvisioned

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	newReplicaCount := req.ReplicaCount
	if req.ReplicaCount == 0 {
		newReplicaCount = volinfo.Subvols[0].ReplicaCount
	}
	if (len(req.Bricks)+len(volinfo.GetBricks()))%(newReplicaCount+volinfo.Subvols[0].ArbiterCount) != 0 {
		return errors.New("invalid number of bricks")
	}

	if volinfo.Type == volume.Replicate || volinfo.Type == volume.DistReplicate {
		if req.ReplicaCount != 0 {
			// TODO: Only considered first sub volume's ReplicaCount
			if req.ReplicaCount < volinfo.Subvols[0].ReplicaCount {
				return errors.New("invalid number of bricks")
			} else if req.ReplicaCount == volinfo.Subvols[0].ReplicaCount {
				return errors.New("replica count is same")
			}
		}
	}

	newBricks, err := volume.NewBrickEntriesFunc(req.Bricks, volinfo.Name, volinfo.VolfileID, volinfo.ID, provisionType)
	if err != nil {
		c.Logger().WithError(err).Error("failed to create new brick entries")
		return err
	}

	if err := c.Set("bricks", newBricks); err != nil {
		return err
	}

	allBricks, err := volume.GetAllBricksInCluster()
	if err != nil {
		return err
	}

	// Used by other peers to check if proposed bricks are already in use.
	// This check is however still prone to races. See issue #314
	if err := c.Set("all-bricks-in-cluster", allBricks); err != nil {
		return err
	}

	checks := brick.PrepareChecks(req.Force, req.Flags)
	if err := c.Set("brick-checks", checks); err != nil {
		return err
	}

	if err := c.Set("newreplicacount", newReplicaCount); err != nil {
		return err
	}

	err = c.Set("volinfo", volinfo)

	return err
}

func startBricksOnExpand(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if volinfo.State != volume.VolStarted {
		return nil
	}

	var newBricks []brick.Brickinfo
	if err := c.Get("bricks", &newBricks); err != nil {
		return err
	}

	// check if bmux is enabled
	bmuxEnabled, err := brickmux.Enabled()
	if err != nil {
		return err
	}

	var allVolumes []*volume.Volinfo
	if bmuxEnabled {
		volumes, err := volume.GetVolumes(context.TODO())
		if err != nil {
			return err
		}
		allVolumes = volumes
	}

	// Start the bricks
	for _, b := range newBricks {

		if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
			continue
		}

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("Starting brick")

		if bmuxEnabled {
			// start multiplexing process
			err := brickmux.Multiplex(b, &volinfo, allVolumes, c.Logger())
			switch err {
			case nil:
				// successfully multiplexed
				continue
			case brickmux.ErrNoCompat:
				// do nothing, fallback to starting a separate process
				c.Logger().WithField("brick", b.String()).Warn(err)
			default:
				return err
			}
		}

		if err := b.StartBrick(c.Logger()); err != nil {
			return err
		}
	}

	return nil
}

func undoStartBricksOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("bricks", &newBricks); err != nil {
		return err
	}

	// Stop the new bricks and delete brick volfile
	for _, b := range newBricks {

		if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
			continue
		}

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("volume expand failed, stopping brick")

		if err := b.StopBrick(c.Logger()); err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"volume": b.VolumeName,
				"brick":  b.String(),
			}).Debug("stopping brick failed")
			// can't know here which of the new bricks started
			// so stopping brick might fail, but log anyway
		}

	}

	return nil
}

func updateVolinfoOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("bricks", &newBricks); err != nil {
		return err
	}

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	var newReplicaCount int
	if err := c.Get("newreplicacount", &newReplicaCount); err != nil {
		return err
	}

	// TODO: Assumption, all subvols are same
	// If New Replica count is different than existing then add one brick to each subvolume
	// Or if the Volume consists of only one subvolume.
	addNewSubvolume := true
	switch volinfo.Subvols[0].Type {
	case volume.SubvolDistribute:
		addNewSubvolume = false
	case volume.SubvolReplicate:
		if newReplicaCount != volinfo.Subvols[0].ReplicaCount {
			addNewSubvolume = false
		}
	default:
		addNewSubvolume = true
	}

	if !addNewSubvolume {
		idx := 0
		for _, b := range newBricks {
			// If number of bricks specified in add brick is more than
			// the number of sub volumes. For example, if number of subvolumes is 2
			// but 4 bricks specified in add brick command.
			if idx >= len(volinfo.Subvols) {
				idx = 0
			}
			volinfo.Subvols[idx].Bricks = append(volinfo.Subvols[idx].Bricks, b)
			idx++
		}
	} else {
		// Create new Sub volumes with given bricks
		subvolIdx := len(volinfo.Subvols)
		bricksCount := newReplicaCount + volinfo.Subvols[0].ArbiterCount
		numSubvols := len(newBricks) / bricksCount
		for i := 0; i < numSubvols; i++ {
			idx := i * bricksCount
			brks := newBricks[idx : idx+bricksCount]
			// If Arbiter count is set then make sure one brick is set
			// as arbiter brick
			if volinfo.Subvols[0].ArbiterCount > 0 {
				arbiterTypeSet := false
				for _, b := range brks {
					if b.Type == brick.Arbiter {
						arbiterTypeSet = true
						break
					}
				}
				if !arbiterTypeSet {
					brks[len(brks)-1].Type = brick.Arbiter
				}
			}
			volinfo.Subvols = append(volinfo.Subvols, volume.Subvol{
				ID:     uuid.NewRandom(),
				Name:   fmt.Sprintf("%s-%s-%d", volinfo.Name, strings.ToLower(volinfo.Subvols[0].Type.String()), subvolIdx),
				Type:   volinfo.Subvols[0].Type,
				Bricks: brks,
			})
			subvolIdx = subvolIdx + 1
		}
	}

	// Update all Subvols Replica count
	for idx := range volinfo.Subvols {
		volinfo.Subvols[idx].ReplicaCount = newReplicaCount
	}

	volinfo.DistCount = len(volinfo.Subvols)

	// update new volinfo in txn ctx
	if err := c.Set("volinfo", volinfo); err != nil {
		return err
	}

	// update new volinfo in etcd store and generate client volfile
	if err := storeVolume(c); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeVolume: failed to store volume info")
		return err
	}

	return nil
}

func resizeLVM(c transaction.TxnCtx) error {
	var req api.VolExpandReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		return err
	}

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	var expansionTpSizePerBrick uint64
	if err := c.Get("expansionTpSizePerBrick", &expansionTpSizePerBrick); err != nil {
		return err
	}

	var expansionMetadataSizePerBrick uint64
	if err := c.Get("expansionMetadataSizePerBrick", &expansionMetadataSizePerBrick); err != nil {
		return err
	}

	var brickVgMapping map[string]string
	if err := c.Get("brickVgMapping", &brickVgMapping); err != nil {
		return err
	}

	if err := expandLocalBricks(volinfo, expansionTpSizePerBrick, expansionMetadataSizePerBrick, brickVgMapping); err != nil {
		return err
	}

	// Update new volume size in bytes.
	volinfo.Capacity = volinfo.Capacity + req.Size
	// update new volinfo in txn ctx
	return c.Set("volinfo", volinfo)

}

// expandLocalBricks expands the local bricks by extending thinpool, metadata pool and lvm for each brick on current node.
func expandLocalBricks(volinfo *volume.Volinfo, expansionTpSizePerBrick uint64, expansionMetadataSizePerBrick uint64, brickVgMapping map[string]string) error {
	for i, sv := range volinfo.Subvols {
		for j, b := range sv.Bricks {
			if uuid.Equal(b.PeerID, gdctx.MyUUID) {
				tpName := fmt.Sprintf("tp_%s_s%d_b%d", volinfo.Name, i+1, j+1)
				lvName := fmt.Sprintf("brick_%s_s%d_b%d", volinfo.Name, i+1, j+1)
				totalExpansionSizePerBrick := expansionTpSizePerBrick + expansionMetadataSizePerBrick

				// extend thinpool
				err := lvmutils.ExtendThinpool(expansionTpSizePerBrick, b.VgName, tpName)
				if err != nil {
					return err
				}
				// extend metadata pool
				err = lvmutils.ExtendMetadataPool(expansionMetadataSizePerBrick, b.VgName, tpName)
				if err != nil {
					return err
				}

				// extend lv
				err = lvmutils.ExtendLV(totalExpansionSizePerBrick, b.VgName, lvName)
				if err != nil {
					return err
				}

				// Update current Vg free size
				err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
