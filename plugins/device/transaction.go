package device

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	log "github.com/sirupsen/logrus"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var peerID string
	var devices []string
	var deviceList []deviceapi.Info

	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerid").Error("Failed to get key from transaction context")
		return err
	}
	if err := c.Get("devices", &devices); err != nil {
		c.Logger().WithError(err).WithField("key", "req").Error("Failed to get key from transaction context")
		return err
	}
	for _, name := range devices {
		tempDevice := deviceapi.Info{
			Name: name,
		}
		deviceList = append(deviceList, tempDevice)
	}
	for index, element := range deviceList {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element.Name)
		if err := pvcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).WithField("device", element.Name).Error("pvcreate failed for device")
			deviceList[index].State = deviceapi.DeviceFailed
			continue
		}
		vgcreateCmd := exec.Command("vgcreate", strings.Replace("vg"+element.Name, "/", "-", -1), element.Name)
		if err := vgcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).WithField("device", element.Name).Error("vgcreate failed for device")
			deviceList[index].State = deviceapi.DeviceFailed
			continue
		}
		deviceList[index].State = deviceapi.DeviceEnabled
	}
	err := AddDevices(deviceList, peerID)
	if err != nil {
		log.WithError(err).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
