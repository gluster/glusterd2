package cmdexec

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
)

func createVgName(device string) string {
	vgName := strings.Replace("vg"+device, "/", "-", -1)
	return vgName
}

// DeviceSetup is used to prepare device before using devices.
func DeviceSetup(c transaction.TxnCtx, device string) error {

	var err error
	defer func() {
		if err != nil {
			DeviceCleanup(c, device)
		}
	}()
	pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", device)
	if err := pvcreateCmd.Run(); err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("pvcreate failed for device")
		return err
	}
	vgName := createVgName(device)
	vgcreateCmd := exec.Command("vgcreate", vgName, device)
	if err = vgcreateCmd.Run(); err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("vgcreate failed for device")
		return err
	}

	return nil

}

// DeviceCleanup is used to clean up devices.
func DeviceCleanup(c transaction.TxnCtx, device string) {
	vgName := createVgName(device)
	vgremoveCmd := exec.Command("vgremove", vgName)
	if err := vgremoveCmd.Run(); err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("vgremove failed for device")
	}
	pvremoveCmd := exec.Command("pvremove", device)
	if err := pvremoveCmd.Run(); err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("pvremove failed for device")
	}
}
