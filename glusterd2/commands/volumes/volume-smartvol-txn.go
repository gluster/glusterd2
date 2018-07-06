package volumecommands

import (
	"os"
	"path"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	config "github.com/spf13/viper"
)

func txnPrepareBricks(c transaction.TxnCtx) error {
	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "req").Debug("Failed to get key from store")
		return err
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {
			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			brickMountDir := path.Dir(b.Path)

			// Create Mount directory
			err := os.MkdirAll(brickMountDir, os.ModeDir|os.ModePerm)
			if err != nil {
				c.Logger().WithError(err).
					WithField("path", brickMountDir).
					Error("Failed to create brick mount directory")
				return err
			}

			// Thin Pool Creation
			err = deviceutils.CreateTP(b.VgName, b.TpName, b.TpSize, b.TpMetadataSize)
			if err != nil {
				c.Logger().WithError(err).
					WithField("vg-name", b.VgName).
					WithField("tp-name", b.TpName).
					WithField("tp-size", b.TpSize).
					WithField("tp-meta-size", b.TpMetadataSize).
					Error("Thinpool Creation failed")
				return err
			}

			// LV Creation
			err = deviceutils.CreateLV(b.VgName, b.TpName, b.LvName, b.Size)
			if err != nil {
				c.Logger().WithError(err).
					WithField("vg-name", b.VgName).
					WithField("tp-name", b.TpName).
					WithField("lv-name", b.LvName).
					WithField("size", b.Size).
					Error("lvcreate failed")
				return err
			}

			dev := "/dev/" + b.VgName + "/" + b.LvName
			// Make Filesystem
			err = deviceutils.MakeXfs(dev)
			if err != nil {
				c.Logger().WithError(err).WithField("dev", dev).Error("mkfs.xfs failed")
				return err
			}

			// Mount the Created FS
			err = deviceutils.BrickMount(dev, brickMountDir)
			if err != nil {
				c.Logger().WithError(err).
					WithField("dev", dev).
					WithField("path", brickMountDir).
					Error("brick mount failed")
				return err
			}

			// Create a directory in Brick Mount
			err = os.MkdirAll(b.Path, os.ModeDir|os.ModePerm)
			if err != nil {
				c.Logger().WithError(err).
					WithField("path", b.Path).
					Error("Failed to create brick directory in mount")
				return err
			}

			// Update current Vg free size
			deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.VgName)

			// Persist mount points in custom fstab file
			// On Glusterd2 restart, all bricks should be
			// mounted using mount -a -T <fstab-file>
			fstabFile := config.GetString("localstatedir") + "/fstab"
			err = deviceutils.FstabAddMount(fstabFile, deviceutils.FstabMount{
				Device:           dev,
				MountPoint:       brickMountDir,
				FilesystemFormat: "xfs",
				MountOptions:     "rw,inode64,noatime,nouuid",
				DumpValue:        "1",
				FsckOption:       "2",
			})
			if err != nil {
				c.Logger().WithError(err).
					WithField("fstab", fstabFile).
					WithField("device", dev).
					WithField("mount", brickMountDir).
					Error("Failed to add entry to fstab file")
				return err
			}
		}
	}

	return nil
}

func txnUndoPrepareBricks(c transaction.TxnCtx) error {
	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "req").Debug("Failed to get key from store")
		return err
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {

			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			brickMountDir := path.Dir(b.Path)

			// UnMount the Brick
			err := deviceutils.BrickUnmount(brickMountDir)
			if err != nil {
				c.Logger().WithError(err).
					WithField("path", brickMountDir).
					Error("brick unmount failed")
			}

			// Remove entry from fstab if available
			fstabFile := config.GetString("localstatedir") + "/fstab"
			err = deviceutils.FstabRemoveMount(fstabFile, brickMountDir)
			if err != nil {
				c.Logger().WithError(err).
					WithField("fstab", fstabFile).
					WithField("mount", brickMountDir).
					Error("Failed to remove entry from fstab file")
			}

			// Remove LV
			err = deviceutils.RemoveLV(b.VgName, b.LvName)
			if err != nil {
				c.Logger().WithError(err).
					WithField("vg-name", b.VgName).
					WithField("lv-name", b.LvName).
					Error("lv remove failed")
			}

			// Remove Thin Pool
			err = deviceutils.RemoveLV(b.VgName, b.TpName)
			if err != nil {
				c.Logger().WithError(err).
					WithField("vg-name", b.VgName).
					WithField("tp-name", b.TpName).
					Error("thin pool remove failed")
			}

			// Update current Vg free size
			deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.VgName)
		}
	}

	return nil
}
