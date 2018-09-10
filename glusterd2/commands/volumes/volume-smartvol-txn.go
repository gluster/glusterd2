package volumecommands

import (
	"errors"
	"os"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	log "github.com/sirupsen/logrus"
)

func txnPrepareBricks(c transaction.TxnCtx) error {
	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField("key", "req").Error("failed to get key from store")
		return err
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {
			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// Create Mount directory
			mountRoot := strings.TrimSuffix(b.Path, b.Mountdir)
			err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm)
			if err != nil {
				c.Logger().WithError(err).WithField("path", mountRoot).Error("failed to create brick mount directory")
				return err
			}

			// Thin Pool Creation
			err = deviceutils.CreateTP(b.VgName, b.TpName, b.TpSize, b.TpMetadataSize)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name":      b.VgName,
					"tp-name":      b.TpName,
					"tp-size":      b.TpSize,
					"tp-meta-size": b.TpMetadataSize,
				}).Error("thinpool creation failed")
				return err
			}

			// LV Creation
			err = deviceutils.CreateLV(b.VgName, b.TpName, b.LvName, b.Size)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name": b.VgName,
					"tp-name": b.TpName,
					"lv-name": b.LvName,
					"size":    b.Size,
				}).Error("lvcreate failed")
				return err
			}

			// Make Filesystem
			err = deviceutils.MakeXfs(b.DevicePath)
			if err != nil {
				c.Logger().WithError(err).WithField("dev", b.DevicePath).Error("mkfs.xfs failed")
				return err
			}

			// Mount the Created FS
			err = deviceutils.BrickMount(b.DevicePath, mountRoot)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"dev":  b.DevicePath,
					"path": mountRoot,
				}).Error("brick mount failed")
				return err
			}

			// Create a directory in Brick Mount
			err = os.MkdirAll(b.Path, os.ModeDir|os.ModePerm)
			if err != nil {
				c.Logger().WithError(err).WithField(
					"path", b.Path).Error("failed to create brick directory in mount")
				return err
			}

			// Update current Vg free size
			err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.VgName)
			if err != nil {
				c.Logger().WithError(err).WithField("vg-name", b.VgName).
					Error("failed to update available size of a device")
				return err
			}
		}
	}

	return nil
}

func txnUndoPrepareBricks(c transaction.TxnCtx) error {
	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField("key", "req").Error("failed to get key from store")
		return err
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {

			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// UnMount the Brick
			mountRoot := strings.TrimSuffix(b.Path, b.Mountdir)
			err := deviceutils.BrickUnmount(mountRoot)
			if err != nil {
				c.Logger().WithError(err).WithField("path", mountRoot).Error("brick unmount failed")
			}

			// Remove LV
			err = deviceutils.RemoveLV(b.VgName, b.LvName)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name": b.VgName,
					"lv-name": b.LvName,
				}).Error("lv remove failed")
			}

			// Remove Thin Pool
			err = deviceutils.RemoveLV(b.VgName, b.TpName)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name": b.VgName,
					"tp-name": b.TpName,
				}).Error("thinpool remove failed")
			}

			// Update current Vg free size
			deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.VgName)
		}
	}

	return nil
}

func txnCleanBricks(c transaction.TxnCtx) error {
	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volinfo").Debug("Failed to get key from store")
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		// UnMount the Brick if mounted
		mountRoot := strings.TrimSuffix(b.Path, b.MountInfo.Mountdir)
		_, err := volume.GetBrickMountInfo(mountRoot)
		if err != nil {
			if !volume.IsMountNotFoundError(err) {
				c.Logger().WithError(err).WithField("path", mountRoot).
					Error("unable to get mount info")
				return err
			}
		} else {
			err := deviceutils.BrickUnmount(mountRoot)
			if err != nil {
				c.Logger().WithError(err).WithField("path", mountRoot).
					Error("brick unmount failed")
				return err
			}
		}

		parts := strings.Split(b.MountInfo.DevicePath, "/")
		if len(parts) != 4 {
			return errors.New("unable to parse device path")
		}
		vgname := parts[2]
		lvname := parts[3]
		tpname, err := deviceutils.GetThinpoolName(vgname, lvname)
		if err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"vg-name": vgname,
				"lv-name": lvname,
			}).Error("failed to get thinpool name")
			return err
		}

		// Remove LV
		err = deviceutils.RemoveLV(vgname, lvname)
		if err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"vg-name": vgname,
				"lv-name": lvname,
			}).Error("lv remove failed")
			return err
		}

		// Remove Thin Pool if LV count is zero, Thinpool will
		// have more LVs in case of snapshots and clones
		numLvs, err := deviceutils.NumberOfLvs(vgname, tpname)
		if err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"vg-name": vgname,
				"tp-name": tpname,
			}).Error("failed to get number of lvs")
			return err
		}

		if numLvs == 0 {
			err = deviceutils.RemoveLV(vgname, tpname)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name": vgname,
					"tp-name": tpname,
				}).Error("thinpool remove failed")
				return err
			}
		}

		// Update current Vg free size
		err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), vgname)
		if err != nil {
			c.Logger().WithError(err).WithField("vg-name", vgname).
				Error("failed to update available size of a device")
			return err
		}
	}

	return nil
}
