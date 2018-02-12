package devicecommands

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/device"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
)

const (
	devicePrefix string = "devices/"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var deviceinfo api.Device

	if err := c.Get("peerid", &deviceinfo.PeerID); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find peer-id")
		return err
	}
	if err := c.Get("device-details", &deviceinfo.Detail); err != nil {
		c.Logger().WithError(err).Error("Failed transaction, cannot find device-details")
		return err
	}
	for index, element := range deviceinfo.Detail {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element.Name)
		if err := pvcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).Error("pvcreate failed for device")
			deviceinfo.Detail[index].State = api.DeviceFailed
			continue
		}
		vgcreateCmd := exec.Command("vgcreate", strings.Replace("vg"+element.Name, "/", "-", -1), element.Name)
		if err := vgcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).Error("vgcreate failed for device")
			deviceinfo.Detail[index].State = api.DeviceFailed
			continue
		}
		deviceinfo.Detail[index].State = api.DeviceEnabled
	}
	deviceDetails, _ := device.GetDevice(deviceinfo.PeerID.String())
	if deviceDetails != nil {
		for _, element := range deviceinfo.Detail {
			deviceDetails.Detail = append(deviceDetails.Detail, element)
		}
		err := device.AddOrUpdateDevice(deviceDetails)
		if err != nil {
			log.WithError(err).Error("Couldn't add deviceinfo to store")
		}
	} else {
		err := device.AddOrUpdateDevice(&deviceinfo)
		if err != nil {
			log.WithError(err).Error("Couldn't add deviceinfo to store")
		}
	}
	return nil
}
