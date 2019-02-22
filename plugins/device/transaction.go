package device

import (
	"os"

	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/fsutils"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/pborman/uuid"
)

func txnPrepareDevice(c oldtransaction.TxnCtx) error {
	var req deviceapi.AddDeviceReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField("key", "req").Error("Failed to get key from transaction context")
		return err
	}
	if req.ProvisionerType == api.ProvisionerTypeLoop {
		return txnPrepareDeviceLoop(req, c)
	}
	return txnPrepareDeviceLvm(req, c)
}

func txnPrepareDeviceLvm(req deviceapi.AddDeviceReq, c oldtransaction.TxnCtx) error {
	var peerID uuid.UUID
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerid").Error("Failed to get key from transaction context")
		return err
	}

	deviceInfo := deviceapi.Info{Device: req.Device}

	err := lvmutils.CreatePV(req.Device)
	if err != nil {
		c.Logger().WithError(err).WithField("device", req.Device).Error("Failed to create physical volume")
		return err
	}
	err = lvmutils.CreateVG(req.Device, deviceInfo.VgName())
	if err != nil {
		c.Logger().WithError(err).WithField("device", req.Device).Error("Failed to create volume group")
		errPV := lvmutils.RemovePV(req.Device)
		if errPV != nil {
			c.Logger().WithError(err).WithField("device", req.Device).Error("Failed to remove physical volume")
		}
		return err
	}

	availableSize, extentSize, err := lvmutils.GetVgAvailableSize(deviceInfo.VgName())
	if err != nil {
		return err
	}
	deviceInfo = deviceapi.Info{
		Device:          req.Device,
		State:           deviceapi.DeviceEnabled,
		AvailableSize:   availableSize,
		TotalSize:       availableSize,
		UsedSize:        0,
		ExtentSize:      extentSize,
		PeerID:          peerID,
		ProvisionerType: api.ProvisionerTypeLvm,
	}

	err = deviceutils.AddOrUpdateDevice(deviceInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}

func txnPrepareDeviceLoop(req deviceapi.AddDeviceReq, c oldtransaction.TxnCtx) error {
	var peerID uuid.UUID
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerid").Error("Failed to get key from transaction context")
		return err
	}

	deviceInfo := deviceapi.Info{Device: req.Device}

	// TODO: Validate Path contains file system and empty

	stat, err := fsutils.StatFs(req.Device)
	if os.IsNotExist(err) {
		c.Logger().WithError(err).WithField("device", req.Device).Error("Failed to prepare device, device not found")
		return err
	}

	deviceInfo = deviceapi.Info{
		Device:          req.Device,
		State:           deviceapi.DeviceEnabled,
		AvailableSize:   stat.Total,
		TotalSize:       stat.Free,
		UsedSize:        stat.Total - stat.Free,
		PeerID:          peerID,
		ProvisionerType: api.ProvisionerTypeLoop,
	}

	err = deviceutils.AddOrUpdateDevice(deviceInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
