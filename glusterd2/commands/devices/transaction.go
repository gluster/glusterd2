package devicecommands

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
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
	for _, element := range deviceinfo.Detail {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element.Name)
		if err := pvcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).Error("pvcreate failed for device")
			element.State = api.DeviceFailed
			continue
		}
		vgcreateCmd := exec.Command("vgcreate", strings.Replace("vg"+element.Name, "/", "-", -1), element.Name)
		if err := vgcreateCmd.Run(); err != nil {
			c.Logger().WithError(err).Error("vgcreate failed for device")
			element.State = api.DeviceFailed
			continue
		}
		element.State = device.DeviceEnabled
	}
	return nil
}
