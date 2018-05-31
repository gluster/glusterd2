package device

import (
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var peerID string
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerid").Error("Failed to get key from transaction context")
		return err
	}

	var device string
	if err := c.Get("device", &device); err != nil {
		c.Logger().WithError(err).WithField("key", "device").Error("Failed to get key from transaction context")
		return err
	}

	var deviceInfo deviceapi.Info

	err := deviceutils.CreatePV(device)
	if err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("Failed to create physical volume")
		return err
	}
	vgName := strings.Replace("vg"+device, "/", "-", -1)
	err = deviceutils.CreateVG(device, vgName)
	if err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("Failed to create volume group")
		errPV := deviceutils.RemovePV(device)
		if errPV != nil {
			c.Logger().WithError(err).WithField("device", device).Error("Failed to remove physical volume")
		}
		return err
	}
	c.Logger().WithField("device", device).Info("Device setup successful, setting device status to 'Enabled'")

	availableSize, extentSize, err := deviceutils.GetVgAvailableSize(vgName)
	if err != nil {
		return err
	}
	deviceInfo = deviceapi.Info{
		Name:          device,
		State:         deviceapi.DeviceEnabled,
		AvailableSize: availableSize,
		ExtentSize:    extentSize,
		PeerID:        peerID,
		VgName:        vgName,
	}

	err = deviceutils.AddDevice(deviceInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
