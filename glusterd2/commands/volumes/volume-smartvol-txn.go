package volumecommands

import (
	"os"
	"path"
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
			var err error
			if req.ProvisionerType == api.ProvisionerTypeLoop {
				err = PrepareBrickLoop(b, c)
			} else {
				err = PrepareBrickLvm(b, c)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// PrepareBrickLvm prepares(Creates thin pool, creates LV, mounts etc.) a single brick
func PrepareBrickLvm(b api.BrickReq, c transaction.TxnCtx) error {
	if b.PeerID != gdctx.MyUUID.String() {
		return nil
	}

	if err := c.Set("freesizeSet."+b.PeerID+b.Path, false); err != nil {
		return err
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
	err = fsutils.Mount(b.DevicePath, mountRoot, b.MntOpts)
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
	err = deviceutils.ReduceDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice, b.TotalSize)
	if err != nil {
		c.Logger().WithError(err).WithField("vg-name", b.VgName).
			Error("failed to update available size of a device")
		return err
	}

	if err := c.Set("freesizeSet."+b.PeerID+b.Path, true); err != nil {
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
	if req.ProvisionerType == api.ProvisionerTypeLoop {
		return txnUndoPrepareBricksLoop(req, c)
	}
	return txnUndoPrepareBricksLvm(req, c)
}

func txnUndoPrepareBricksLvm(req api.VolCreateReq, c transaction.TxnCtx) error {
	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {

			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// UnMount the Brick
			mountRoot := strings.TrimSuffix(b.Path, b.BrickDirSuffix)
			err := fsutils.Unmount(mountRoot)
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

			var freesizeSet bool
			key := "freesizeSet." + b.PeerID + b.Path
			if err := c.Get(key, &freesizeSet); err != nil {
				c.Logger().WithError(err).WithField("key", key).Error("failed to get key from store")
				return err
			}

			// Reset Free size only if freeSize is set in transaction
			if freesizeSet {
				deviceutils.AddDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice, b.TotalSize)
			}
		}
	}

	return nil
}

// PrepareBrickLoop prepares a single brick
func PrepareBrickLoop(b api.BrickReq, c transaction.TxnCtx) error {
	if b.PeerID != gdctx.MyUUID.String() {
		return nil
	}

	if err := c.Set("freesizeSet."+b.PeerID+b.Path, false); err != nil {
		return err
	}

	// Create Mount directory
	mountRoot := strings.TrimSuffix(b.Path, b.BrickDirSuffix)
	err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm)
	if err != nil {
		c.Logger().WithError(err).WithField("path", mountRoot).Error("failed to create brick mount directory")
		return err
	}

	// Dir creation
	err = os.MkdirAll(path.Dir(b.DevicePath), 0700)
	if err != nil {
		c.Logger().WithError(err).WithField("device", path.Dir(b.DevicePath)).Error("brick device directory create failed")
		return err
	}

	// Loop file creation
	loopFile, err := os.OpenFile(b.DevicePath, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		c.Logger().WithError(err).WithField("device", b.DevicePath).Error("brick device create failed")
		return err
	}
	defer loopFile.Close()
	err = os.Truncate(b.DevicePath, int64(b.Size))
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"device": b.DevicePath,
			"size":   b.Size,
		}).Error("brick device truncate failed")
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
	err = fsutils.Mount(b.DevicePath, mountRoot, b.MntOpts)
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
	err = deviceutils.ReduceDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice, b.TotalSize)
	if err != nil {
		c.Logger().WithError(err).WithField("vg-name", b.VgName).
			Error("failed to update available size of a device")
		return err
	}

	if err := c.Set("freesizeSet."+b.PeerID+b.Path, true); err != nil {
		return err
	}

	return nil
}

func txnUndoPrepareBricksLoop(req api.VolCreateReq, c transaction.TxnCtx) error {
	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {

			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// UnMount the Brick
			mountRoot := strings.TrimSuffix(b.Path, b.BrickDirSuffix)
			err := fsutils.Unmount(mountRoot)
			if err != nil {
				c.Logger().WithError(err).WithField("path", mountRoot).Error("brick unmount failed")
			}

			// Remove Loop device
			err = os.Remove(b.DevicePath)
			if err != nil {
				c.Logger().WithError(err).WithField("device", b.DevicePath).Error("brick device remove failed")
			}

			// Remove Thin Pool
			err = os.RemoveAll(path.Dir(b.DevicePath))
			if err != nil {
				c.Logger().WithError(err).WithField("device", path.Dir(b.DevicePath)).Error("brick device directory remove failed")
			}

			var freesizeSet bool
			key := "freesizeSet." + b.PeerID + b.Path
			if err := c.Get(key, &freesizeSet); err != nil {
				c.Logger().WithError(err).WithField("key", key).Error("failed to get key from store")
				return err
			}

			// Reset Free size only if freeSize is set in transaction
			if freesizeSet {
				deviceutils.AddDeviceFreeSize(gdctx.MyUUID.String(), b.RootDevice, b.TotalSize)
			}
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

	if volinfo.ProvisionerType == api.ProvisionerTypeLoop {
		return volume.CleanBricksLoop(&volinfo)
	}

	return volume.CleanBricksLvm(&volinfo)
}
