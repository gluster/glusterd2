package deviceutils

import (
	"os/exec"
)

//CreatePV is used to create physical volume.
func CreatePV(device string) error {
	pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", device)
	return pvcreateCmd.Run()
}

//CreateVG is used to create volume group
func CreateVG(device string, vgName string) error {
	vgcreateCmd := exec.Command("vgcreate", vgName, device)
	return vgcreateCmd.Run()
}

//RemoveVG is used to remove volume group.
func RemoveVG(vgName string) error {
	vgremoveCmd := exec.Command("vgremove", vgName)
	return vgremoveCmd.Run()
}

//RemovePV is used to remove physical volume
func RemovePV(device string) error {
	pvremoveCmd := exec.Command("pvremove", device)
	return pvremoveCmd.Run()
}
