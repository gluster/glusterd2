package device

import (
	"github.com/gluster/glusterd2/glusterd2/transaction"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/cmdexec"
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

	for index, device := range deviceList {
		err := cmdexec.DeviceSetup(device.Name)
		if err != nil {
			deviceList[index].State = deviceapi.DeviceFailed
			continue
		}
		deviceList[index].State = deviceapi.DeviceEnabled
	}

	err := AddDevices(deviceList, peerID)
	if err != nil {
		c.Logger().WithError(err).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
