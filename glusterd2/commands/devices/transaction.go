package devicecommands

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/device"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var peerID uuid.UUID
	var req api.AddDeviceReq
	var deviceList []api.DeviceInfo
	if err := c.Get("peerid", peerID); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find peer-id")
		return err
	}
	if err := c.Get("device-details", req); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find device-details")
		return err
	}
	for _, name := range req.Devices {
		tempDevice := api.DeviceInfo{
			Name: name,
		}
		deviceList = append(deviceList, tempDevice)
	}
	for index, element := range deviceList {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element.Name)
		if err := pvcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).WithField("device", element.Name).Error("pvcreate failed for device")
			deviceList[index].State = api.DeviceFailed
			continue
		}
		vgcreateCmd := exec.Command("vgcreate", strings.Replace("vg"+element.Name, "/", "-", -1), element.Name)
		if err := vgcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).WithField("device", element.Name).Error("vgcreate failed for device")
			deviceList[index].State = api.DeviceFailed
			continue
		}
		deviceList[index].State = api.DeviceEnabled
	}
	err := device.AddDevices(deviceList, peerID.String())
	if err != nil {
		log.WithError(err).Error("Couldn't add deviceinfo to store")
	}
	return nil
}
