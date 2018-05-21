package deviceutils

import (
	"os/exec"
)

//
func CreatePV(device string) error {

	pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", device)
	if err := pvcreateCmd.Run(); err != nil {
		return err
	}
	return nil
}

//
func CreateVG(device string, vgName string) error {

	vgcreateCmd := exec.Command("vgcreate", vgName, device)
	if err := vgcreateCmd.Run(); err != nil {
		return err
	}
	return nil
}

//
func RemoveVG(vgName string) error {
	vgremoveCmd := exec.Command("vgremove", vgName)
	if err := vgremoveCmd.Run(); err != nil {
		return err
	}
	return nil
}

//
func RemovePV(device string) error {
	pvremoveCmd := exec.Command("pvremove", device)
	if err := pvremoveCmd.Run(); err != nil {
		return err
	}
	return nil
}
