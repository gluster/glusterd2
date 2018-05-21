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

	var devices []string
	if err := c.Get("devices", &devices); err != nil {
		c.Logger().WithError(err).WithField("key", "req").Error("Failed to get key from transaction context")
		return err
	}

	var deviceList []deviceapi.Info
	for _, name := range devices {
		tempDevice := deviceapi.Info{
			Name: name,
		}
		deviceList = append(deviceList, tempDevice)
	}

	var failedDevice []string
	var successDevice []deviceapi.Info
	for index, device := range deviceList {
		err := deviceutils.CreatePV(device.Name)
		if err != nil {
			c.Logger().WithError(err).WithField("device", device.Name).Error("Failed to create physical volume")
			continue
		}
		vgName := strings.Replace("vg"+device.Name, "/", "-", -1)
		err = deviceutils.CreateVG(device.Name, vgName)
		if err != nil {
			c.Logger().WithError(err).WithField("device", device.Name).Error("Failed to create volume group")
			err = deviceutils.RemovePV(device.Name)
			if err != nil {
				c.Logger().WithError(err).WithField("device", device.Name).Error("Failed to remove physical volume")
				failedDevice = append(failedDevice, device.Name)
			}
		}
		c.Logger().WithError(err).WithField("device", device.Name).Error("Setup device successful, setting device status to 'DeviceEnabled'")
		deviceList[index].State = deviceapi.DeviceEnabled
		successDevice = append(successDevice, deviceList[index])
	}

	err := AddDevices(successDevice, peerID)
	if err != nil {
		c.Logger().WithError(err).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
