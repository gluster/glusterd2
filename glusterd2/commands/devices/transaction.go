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
	var devices []api.DeviceInfo

	if err := c.Get("peerid", peerID); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find peer-id")
		return err
	}
	if err := c.Get("device-details", &devices); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find device-details")
		return err
	}
	for index, element := range devices {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element.Name)
		if err := pvcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).Error("pvcreate failed for device")
			devices[index].State = api.DeviceFailed
			continue
		}
		vgcreateCmd := exec.Command("vgcreate", strings.Replace("vg"+element.Name, "/", "-", -1), element.Name)
		if err := vgcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).Error("vgcreate failed for device")
			devices[index].State = api.DeviceFailed
			continue
		}
		devices[index].State = api.DeviceEnabled
	}
	deviceDetails, _ := device.GetDevice(peerID.String())
	if deviceDetails != nil {
		for _, element := range deviceDetails {
			deviceDetails = append(deviceDetails, element)
		}
		err := device.AddOrUpdateDevice(deviceDetails, peerID.String())
		if err != nil {
			log.WithError(err).Error("Couldn't add deviceinfo to store")
		}
	} else {
		err := device.AddOrUpdateDevice(devices, peerID.String())
		if err != nil {
			log.WithError(err).Error("Couldn't add deviceinfo to store")
		}
	}
	return nil
}
