package cmdexec

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func createVgName(device string) string {
	vgName := strings.Replace("vg"+device, "/", "-", -1)
	return vgName
}

func DeviceSetup(device string) error {

	var err error
	pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", device)
	if err := pvcreateCmd.Run(); err != nil {
		log.WithError(err).WithField("device", device).Error("pvcreate failed for device")
		return err
	}
	vgName := createVgName(device)
	vgcreateCmd := exec.Command("vgcreate", vgName, device)
	if err = vgcreateCmd.Run(); err != nil {
		log.WithError(err).WithField("device", device).Error("vgcreate failed for device")
		return err
	}

	defer func() {
		if err != nil {
			DeleteDevice(device)
		}
	}()
	return nil

}

func DeleteDevice(device string) {
	vgName := createVgName(device)
	vgremoveCmd := exec.Command("vgremove", vgName)
	if err := vgremoveCmd.Run(); err != nil {
		log.WithError(err).WithField("device", device).Error("vgremove failed for device")
	}
	pvremoveCmd := exec.Command("pvremove", device)
	if err := pvremoveCmd.Run(); err != nil {
		log.WithError(err).WithField("device", device).Error("pvremove failed for device")
	}
}
