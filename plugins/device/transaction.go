package device

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var peerID uuid.UUID
	var req deviceapi.AddDeviceReq
	var deviceList []deviceapi.Info
	if err := c.Get("peerid", peerID); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find peer-id")
		return err
	}
	if err := c.Get("req", req); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find device-details")
		return err
	}
	for _, name := range req.Devices {
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
	err := AddDevices(deviceList, peerID.String())
	if err != nil {
		log.WithError(err).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
