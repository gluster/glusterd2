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

// DeviceSetup is used to prepare device before using devices.
func DeviceSetup(device string) error {

	var err error
	defer func() {
		if err != nil {
			DeviceCleanup(device)
		}
	}()
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

	return nil

}

// DeviceCleanup is used to clean up devices.
func DeviceCleanup(device string) {
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
