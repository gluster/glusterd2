package snapshotcommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const volumeIDXattrKey = "trusted.glusterfs.volume-id"

func snapRestore(c transaction.TxnCtx) error {
	var snapname string
	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapInfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}
	snapVol := &snapInfo.SnapVolinfo

	volinfo, err := volume.GetVolume(snapInfo.ParentVolume)
	if err != nil {
		return err
	}

	//Unlike snapshot volumes, bricks of the regular volume
	//should be mounted regardless of the status.
	//So terminating the brick process keeping the mount point
	onlineBricks, err := snapshot.GetOnlineBricks(snapVol)
	if err != nil {
		return err
	}
	offlineBricks, err := snapshot.GetOfflineBricks(snapVol)
	if err != nil {
		return err
	}

	//Do a proper snapshot stop, once there is a generic way of stopping all the proceess of a volume.
	for _, b := range onlineBricks {
		if err = b.TerminateBrick(); err != nil {
			if err = b.StopBrick(c.Logger()); err != nil {
				return err
			}
		}
		if err := unix.Setxattr(b.Path, volumeIDXattrKey, []byte(volinfo.ID), unix.XATTR_REPLACE); err != nil {
			return err
		}
	}

	mtab, err := volume.GetMounts()
	if err != nil {
		return err
	}

	for _, b := range offlineBricks {
		if err := volume.MountBrickDirectory(snapVol, &b, mtab); err != nil {
			return err
		}
		if err := unix.Setxattr(b.Path, volumeIDXattrKey, []byte(volinfo.ID), unix.XATTR_REPLACE); err != nil {
			return err
		}
	}

	//TODO Stop other process of snapshot volume

	return nil
}

func remountBrick(b brick.Brickinfo, volinfo *volume.Volinfo, mtab []*volume.Mntent) error {
	if err := volume.UmountBrick(b); err != nil {
		return err
	}
	err := volume.MountBrickDirectory(volinfo, &b, mtab)
	return err

}

func undoSnapStore(c transaction.TxnCtx) error {
	var snapInfo snapshot.Snapinfo
	var volinfo volume.Volinfo

	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := snapshot.AddOrUpdateSnapFunc(&snapInfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeSnapshot: failed to store snapshot info")
		return err
	}

	// Regenerate the volfile of original volume
	err := volgen.VolumeVolfileToStore(&volinfo, volinfo.Name, "client")
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"template": "client",
			"volfile":  volinfo.Name,
		}).Error("failed to generate volfile and save to store")
		return err
	}

	// Regenerate the Volfile of snapshot Volume
	err = volgen.VolumeVolfileToStore(&snapInfo.SnapVolinfo, snapInfo.SnapVolinfo.Name, "client")
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"template": "client",
			"volfile":  snapInfo.SnapVolinfo.Name,
		}).Error("failed to generate snapshot volfile and save to store")
		return err
	}

	// Generate brick Volfiles if Snapshot was started earlier
	if snapInfo.SnapVolinfo.State == volume.VolStarted {
		err = volgen.GenerateBricksVolfiles(&snapInfo.SnapVolinfo, snapInfo.SnapVolinfo.GetLocalBricks())
		if err != nil {
			return err
		}
	}

	if err = volume.AddOrUpdateVolumeFunc(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("failed to store volume info")
		return err
	}

	return nil
}

func undoSnapRestore(c transaction.TxnCtx) error {
	var snapname string
	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapInfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}
	snapVol := &snapInfo.SnapVolinfo

	mtab, err := volume.GetMounts()
	if err != nil {
		return err
	}

	for _, b := range snapVol.GetLocalBricks() {

		if err := remountBrick(b, snapVol, mtab); err != nil {
			return err
		}
		if snapVol.State == volume.VolStarted {
			if err := b.StartBrick(c.Logger()); err != nil {
				if err == errors.ErrProcessAlreadyRunning {
					continue
				}
				return err
			}
		} else {
			if err = b.TerminateBrick(); err != nil {
				if err = b.StopBrick(c.Logger()); err != nil {
					//Process might not be running,
					//TODO once we have errors.ErrProcessAlreadyStopped
					//check for other errors
				}
				if err := volume.UmountBrick(b); err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func createVolumeBrickFromSnap(bricks []brick.Brickinfo, vol *volume.Volinfo) []brick.Brickinfo {
	var newBricks []brick.Brickinfo
	for _, b := range bricks {
		newBrick := brick.Brickinfo{
			Decommissioned: b.Decommissioned,
			Hostname:       b.Hostname,
			ID:             b.ID,
			MountInfo:      b.MountInfo,
			Path:           b.Path,
			PeerID:         b.PeerID,
			Type:           b.Type,
			VolumeID:       vol.ID,
			VolumeName:     vol.Name,
			VolfileID:      vol.Name,
			PType:          b.PType,
		}
		newBricks = append(newBricks, newBrick)
	}
	return newBricks
}

func createRestoreVolinfo(snapinfo *snapshot.Snapinfo, vol *volume.Volinfo) volume.Volinfo {
	var newVol volume.Volinfo
	snapVol := &snapinfo.SnapVolinfo

	//Should this be snap auth or original vol auth?
	newVol.Auth = vol.Auth
	newVol.DistCount = snapVol.DistCount
	newVol.GraphMap = snapVol.GraphMap
	newVol.ID = vol.ID
	newVol.Metadata = snapVol.Metadata
	newVol.Name = snapinfo.ParentVolume
	newVol.Options = snapVol.Options
	for key, value := range snapinfo.OptionChange {
		newVol.Options[key] = value
	}

	newVol.SnapList = vol.SnapList
	newVol.State = vol.State
	newVol.Transport = snapVol.Transport
	newVol.Type = snapVol.Type
	newVol.VolfileID = newVol.Name
	for idx, subvol := range snapVol.Subvols {
		subvolType := volume.SubvolTypeToString(subvol.Type)
		name := fmt.Sprintf("%s-%s-%d", vol.Name, strings.ToLower(subvolType), idx)
		bricks := createVolumeBrickFromSnap(subvol.Bricks, vol)
		s := volume.Subvol{
			ArbiterCount:    subvol.ArbiterCount,
			DisperseCount:   subvol.ArbiterCount,
			ID:              subvol.ID,
			Name:            name,
			RedundancyCount: subvol.RedundancyCount,
			ReplicaCount:    subvol.ReplicaCount,
			Type:            subvol.Type,
			Subvols:         subvol.Subvols,
			Bricks:          bricks,
		}
		newVol.Subvols = append(newVol.Subvols, s)
		/*
			TODO
			*Checksum
			*newVol.Version = snapVol.Version
		*/
	}
	return newVol
}

func storeSnapRestore(c transaction.TxnCtx) error {
	var snapname string
	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapInfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}
	snapVol := &snapInfo.SnapVolinfo

	vol, err := volume.GetVolume(snapInfo.ParentVolume)
	if err != nil {
		return err
	}

	newVolinfo := createRestoreVolinfo(snapInfo, vol)

	// Volfile of restored volume
	err = volgen.VolumeVolfileToStore(&newVolinfo, newVolinfo.Name, "client")
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"template": "client",
			"volfile":  newVolinfo.Name,
		}).Error("failed to generate volfile and save to store")
		return err
	}

	if err := volume.AddOrUpdateVolumeFunc(&newVolinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", newVolinfo.Name).Error("failed to store volume info")
		return err
	}

	if err := snapshot.DeleteSnapshot(snapInfo); err != nil {
		c.Logger().WithError(err).WithField(
			"snapshot", snapVol.Name).Error("failed to delete snap from store")
		return err
	}
	if err := volgen.DeleteVolfiles(snapVol.VolfileID); err != nil {
		c.Logger().WithError(err).
			WithField("snapshot", snapshot.GetStorePath(snapInfo)).
			Warn("failed to delete volfiles of snapshot")
		return err
	}

	// Delete Snapshot Brick volfiles if Snapvol was in started state
	if snapVol.State == volume.VolStarted {
		if err := volgen.DeleteBricksVolfiles(snapVol.GetLocalBricks()); err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"template": "brick",
				"volume":   snapVol.Name,
			}).Error("failed to delete brick volfiles")
			return err
		}
	}

	return nil
}

func cleanParentBricks(c transaction.TxnCtx) error {
	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volinfo").Debug("Failed to get key from store")
		return err
	}

	return volume.CleanBricks(&volinfo)
}

func registerSnapRestoreStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"snap-restore.Commit", snapRestore},
		{"snap-restore.UndoCommit", undoSnapRestore},
		{"snap-restore.UndoStore", undoSnapStore},
		{"snap-restore.Store", storeSnapRestore},
		{"snap-restore.CleanBricks", cleanParentBricks},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func snapshotRestoreHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	snapname := mux.Vars(r)["snapname"]

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, snapname, snapinfo.ParentVolume)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	//Fetching snapinfo again, but this time inside a lock
	snapinfo, err = snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	snapvolinfo := &snapinfo.SnapVolinfo

	vol, err := volume.GetVolume(snapinfo.ParentVolume)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	if vol.State == volume.VolStarted {
		errMsg := fmt.Sprintf("Volume %s must be in stopped state before restoring.", vol.Name)
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	bricksAutoProvisioned := vol.IsAutoProvisioned() || vol.IsSnapshotProvisioned()
	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "snap-restore.Commit",
			UndoFunc: "snap-restore.UndoCommit",
			Nodes:    snapvolinfo.Nodes(),
		},
		{
			DoFunc:   "snap-restore.Store",
			UndoFunc: "snap-restore.UndoStore",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "snap-restore.CleanBricks",
			Nodes:  vol.Nodes(),
			Skip:   !bricksAutoProvisioned,
		},
	}
	if err = txn.Ctx.Set("snapname", snapname); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err = txn.Ctx.Set("snapinfo", snapinfo); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Ctx.Set("volinfo", vol); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("snapshot restore transaction failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	msg := fmt.Sprintf("Snapshot %s restored to volume %s", snapvolinfo.Name, vol.Name)
	txn.Ctx.Logger().WithField("snapshot", snapname).Info(msg)

	//Get the updated volinfo
	vol, err = volume.GetVolume(vol.Name)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	resp := volume.CreateVolumeInfoResp(vol)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)

}
