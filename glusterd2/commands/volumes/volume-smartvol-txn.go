package volumecommands

import (
	"os"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	log "github.com/sirupsen/logrus"
)

func txnPrepareBricks(c transaction.TxnCtx) error {
	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithFields(log.Fields{
			"error": err,
			"key":   "req",
		}).Error("failed to get key from store")
		return err
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {
			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// Create Mount directory
			err := os.MkdirAll(b.Mountdir, os.ModeDir|os.ModePerm)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error": err,
					"path":  b.Mountdir,
				}).Error("failed to create brick mount directory")
				return err
			}

			// Thin Pool Creation
			err = deviceutils.CreateTP(b.VgName, b.TpName, b.TpSize, b.TpMetadataSize)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error":        err,
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
				c.Logger().WithFields(log.Fields{
					"error":   err,
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
			err = deviceutils.BrickMount(b.DevicePath, b.Mountdir)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error": err,
					"dev":   b.DevicePath,
					"path":  b.Mountdir,
				}).Error("brick mount failed")
				return err
			}

			// Create a directory in Brick Mount
			err = os.MkdirAll(b.Path, os.ModeDir|os.ModePerm)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error": err,
					"path":  b.Path,
				}).Error("failed to create brick directory in mount")
				return err
			}

			// Update current Vg free size
			deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.VgName)
		}
	}

	return nil
}

func txnUndoPrepareBricks(c transaction.TxnCtx) error {
	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithFields(log.Fields{
			"error": err,
			"key":   "req",
		}).Error("failed to get key from store")
		return err
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {

			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// UnMount the Brick
			err := deviceutils.BrickUnmount(b.Mountdir)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error": err,
					"path":  b.Mountdir,
				}).Error("brick unmount failed")
			}

			// Remove LV
			err = deviceutils.RemoveLV(b.VgName, b.LvName)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error":   err,
					"vg-name": b.VgName,
					"lv-name": b.LvName,
				}).Error("lv remove failed")
			}

			// Remove Thin Pool
			err = deviceutils.RemoveLV(b.VgName, b.TpName)
			if err != nil {
				c.Logger().WithFields(log.Fields{
					"error":   err,
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
