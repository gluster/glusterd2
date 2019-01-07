package volumecommands

import (
	"os"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/fsutils"
	"github.com/gluster/glusterd2/pkg/lvmutils"
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
			err := PrepareBrick(b, c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// PrepareBrick prepares(Creates thin pool, creates LV, mounts etc.) a single brick
func PrepareBrick(b api.BrickReq, c transaction.TxnCtx) error {
	if b.PeerID != gdctx.MyUUID.String() {
		return nil
	}

	// Create Mount directory
	mountRoot := strings.TrimSuffix(b.Path, b.BrickDirSuffix)
	err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm)
	if err != nil {
		c.Logger().WithError(err).WithField("path", mountRoot).Error("failed to create brick mount directory")
		return err
	}

	// Thin Pool Creation
	err = lvmutils.CreateTP(b.VgName, b.TpName, b.TpSize, b.TpMetadataSize)
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
	err = lvmutils.CreateLV(b.VgName, b.TpName, b.LvName, b.Size)
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
	var mkfsOpts []string
	if b.Type == "arbiter" {
		mkfsOpts = []string{"-i", "size=512", "-n", "size=8192", "-i", "maxpct=0"}
	} else {
		mkfsOpts = []string{"-i", "size=512", "-n", "size=8192"}
	}
	err = fsutils.MakeXfs(b.DevicePath, mkfsOpts...)
	if err != nil {
		c.Logger().WithError(err).WithField("dev", b.DevicePath).Error("mkfs.xfs failed")
		return err
	}

	// Mount the Created FS
	err = lvmutils.MountLV(b.DevicePath, mountRoot, b.MntOpts)
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
	err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice)
	if err != nil {
		c.Logger().WithError(err).WithField("vg-name", b.VgName).
			Error("failed to update available size of a device")
		return err
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
			mountRoot := strings.TrimSuffix(b.Path, b.BrickDirSuffix)
			err := lvmutils.UnmountLV(mountRoot)
			if err != nil {
				c.Logger().WithError(err).WithField("path", mountRoot).Error("brick unmount failed")
			}

			// Remove LV
			err = lvmutils.RemoveLV(b.VgName, b.LvName, true)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name": b.VgName,
					"lv-name": b.LvName,
				}).Error("lv remove failed")
			}

			// Remove Thin Pool
			err = lvmutils.RemoveLV(b.VgName, b.TpName, true)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"vg-name": b.VgName,
					"tp-name": b.TpName,
				}).Error("thinpool remove failed")
			}

			// Update current Vg free size
			deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice)
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

	return volume.CleanBricks(&volinfo)
}
