package volumecommands

import (
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/provisioners"
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

	var provisioner provisioners.Provisioner
	var err error
	if req.Provisioner == "" {
		provisioner = provisioners.GetDefault()
	} else {
		provisioner, err = provisioners.Get(req.Provisioner)
		if err != nil {
			c.Logger().WithError(err).WithField("name", req.Provisioner).Error("invalid provisioner")
			return err
		}
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {
			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			err := provisioner.CreateBrick(b.Device, b.Name, b.Size, req.SnapshotReserveFactor)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"device":           b.Device,
					"name":             b.Name,
					"size":             b.Size,
					"snapshot-reserve": req.SnapshotReserveFactor,
				}).Error("brick creation failed")
				return err
			}

			err = provisioner.CreateBrickFS(b.Device, b.Name, "xfs")
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"device": b.Device,
					"fstype": "xfs",
				}).Error("create brick filesystem failed")
				return err
			}

			// Mount the Created FS
			err = provisioner.MountBrick(b.Device, b.Name, b.Path)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"device": b.Device,
					"path":   b.Path,
					"name":   b.Name,
				}).Error("brick mount failed")
				return err
			}

			// Create a directory in Brick Mount
			err = provisioner.CreateBrickDir(b.Path)
			if err != nil {
				c.Logger().WithError(err).WithField(
					"path", b.Path).Error("failed to create brick directory in mount")
				return err
			}

			availableSize, extentSize, err := provisioner.AvailableSize(b.Device)
			if err != nil {
				c.Logger().WithError(err).WithField("device", b.Device).
					Error("failed to get available size of a device")
				return err
			}
			err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.Device, availableSize, extentSize)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"peerid":        gdctx.MyUUID.String(),
					"device":        b.Device,
					"availablesize": availableSize,
				}).Error("failed to update available size of a device")
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

	var provisioner provisioners.Provisioner
	var err error
	if req.Provisioner == "" {
		provisioner = provisioners.GetDefault()
	} else {
		provisioner, err = provisioners.Get(req.Provisioner)
		if err != nil {
			c.Logger().WithError(err).WithField("name", req.Provisioner).Error("invalid provisioner")
			return err
		}
	}

	for _, sv := range req.Subvols {
		for _, b := range sv.Bricks {

			if b.PeerID != gdctx.MyUUID.String() {
				continue
			}

			// UnMount the Brick
			err := provisioner.UnmountBrick(b.Path)
			if err != nil {
				c.Logger().WithError(err).WithField("path", b.Path).Error("brick unmount failed")
			}

			// Remove Brick
			err = provisioner.RemoveBrick(b.Device, b.Name)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"device": b.Device,
					"name":   b.Name,
				}).Error("lv remove failed")
			}

			availableSize, extentSize, err := provisioner.AvailableSize(b.Device)
			if err != nil {
				c.Logger().WithError(err).WithField("device", b.Device).
					Error("failed to get available size of a device")
				return err
			}
			err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.Device, availableSize, extentSize)
			if err != nil {
				c.Logger().WithError(err).WithFields(log.Fields{
					"peerid":        gdctx.MyUUID.String(),
					"device":        b.Device,
					"availablesize": availableSize,
				}).Error("failed to update available size of a device")
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

	var provisioner provisioners.Provisioner
	var err error
	if volinfo.Provisioner == "" {
		provisioner = provisioners.GetDefault()
	} else {
		provisioner, err = provisioners.Get(volinfo.Provisioner)
		if err != nil {
			c.Logger().WithError(err).WithField("name", volinfo.Provisioner).Error("invalid provisioner")
			return err
		}
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
			err := provisioner.UnmountBrick(b.Path)
			if err != nil {
				c.Logger().WithError(err).WithField("path", b.Path).Error("brick unmount failed")
				return err
			}
		}

		err = provisioner.RemoveBrick(b.Device, b.Name)
		if err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"device": b.Device,
				"name":   b.Name,
			}).Error("remove brick failed")
		}

		availableSize, extentSize, err := provisioner.AvailableSize(b.Device)
		if err != nil {
			c.Logger().WithError(err).WithField("device", b.Device).
				Error("failed to get available size of a device")
			return err
		}
		err = deviceutils.UpdateDeviceFreeSize(gdctx.MyUUID.String(), b.Device, availableSize, extentSize)
		if err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"peerid":        gdctx.MyUUID.String(),
				"device":        b.Device,
				"availablesize": availableSize,
			}).Error("failed to update available size of a device")
			return err
		}
	}

	return nil
}
