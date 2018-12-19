package device

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/pborman/uuid"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var peerID uuid.UUID
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerid").Error("Failed to get key from transaction context")
		return err
	}

	var device string
	if err := c.Get("device", &device); err != nil {
		c.Logger().WithError(err).WithField("key", "device").Error("Failed to get key from transaction context")
		return err
	}

	deviceInfo := deviceapi.Info{Device: device}

	err := lvmutils.CreatePV(device)
	if err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("Failed to create physical volume")
		return err
	}
	err = lvmutils.CreateVG(device, deviceInfo.VgName())
	if err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("Failed to create volume group")
		errPV := lvmutils.RemovePV(device)
		if errPV != nil {
			c.Logger().WithError(err).WithField("device", device).Error("Failed to remove physical volume")
		}
		return err
	}
	c.Logger().WithField("device", device).Info("Device setup successful, setting device status to 'Enabled'")

	availableSize, extentSize, err := lvmutils.GetVgAvailableSize(deviceInfo.VgName())
	if err != nil {
		return err
	}
	deviceInfo = deviceapi.Info{
		Device:        device,
		State:         deviceapi.DeviceEnabled,
		AvailableSize: availableSize,
		TotalSize:     availableSize,
		UsedSize:      0,
		ExtentSize:    extentSize,
		PeerID:        peerID,
	}

	err = deviceutils.AddOrUpdateDevice(deviceInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}

func txnDeleteDevice(c transaction.TxnCtx) error {
	var err error
	var peerID string

	if err = c.Get("peerid", &peerID); err != nil {
		return err
	}

	var deviceName string
	if err = c.Get("device", &deviceName); err != nil {
		return err
	}

	devices, err := deviceutils.GetDevices(peerID)
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return gderrors.ErrDeviceNotFound
	}

	vgName := ""
	for _, device := range devices {
		if device.Device == deviceName {
			vgName = device.VgName()
		}
	}

	if vgName == "" {
		return gderrors.ErrDeviceNameNotFound
	}

	nlvs, err := lvmutils.NumberOfLvs(vgName, "")
	if err != nil {
		c.Logger().WithError(err).WithField("device", deviceName).Error("Failed to get number of Lvs")
		return err
	}

	if nlvs > 0 {
		c.Logger().WithError(err).WithField("device", deviceName).Error("Cannot remove device, volume exists on device")
		return errors.New("Cannot remove device, volumes exists on device")
	}

	// Remove VG
	if err = lvmutils.RemoveVG(vgName); err != nil {
		c.Logger().WithError(err).WithField("device", deviceName).Error("Failed to remove volume group")
		return err
	}

	//Remove PV
	if err = lvmutils.RemovePV(deviceName); err != nil {
		c.Logger().WithError(err).WithField("device", deviceName).Error("Failed to remove physical volume")
		return err
	}

	err = deviceutils.DeleteDevice(peerID, deviceName)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Failed to remove device")
		return err
	}
	return nil
}
