package lvm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
)

const (
	//CreateCommand is path to lvm create
	CreateCommand string = "/sbin/lvcreate"
	//RemoveCommand is path to lvm create
	RemoveCommand string = "/sbin/lvremove"
	//PvCreateCommand is path to lvm create
	PvCreateCommand string = "/sbin/pvcreate"
	//VgCreateCommand is path to lvm create
	VgCreateCommand string = "/sbin/vgcreate"
)

// CommonPrevalidation checks the lvm related validation for snapshot
func CommonPrevalidation(lvmCommand string) error {
	fileInfo, err := os.Stat(lvmCommand)
	if err != nil {
		return err
	}
	switch mode := fileInfo.Mode(); {

	case mode.IsRegular() == false:
		return err
	case mode&0111 == 0:
		return err
	}
	return nil
}

//IsThinLV check for lvm compatibility for a path
func IsThinLV(brickPath string) bool {
	mountRoot, err := volume.GetBrickMountRoot(brickPath)
	if err != nil {
		return false
	}

	mntInfo, err := volume.GetBrickMountInfo(mountRoot)
	if err != nil {
		return false
	}

	out, err := utils.ExecuteCommandOutput("/sbin/lvs", "--noheadings", "-o", "pool_lv", mntInfo.FsName)
	if err != nil {
		return false
	}

	thinLV := strings.TrimSpace(string(out))

	if thinLV == "" {
		return false
	}
	return true
}

//MountSnapshotDirectory will mount the snapshot bricks to the given path
func MountSnapshotDirectory(mountPath string, mountData brick.MountInfo) error {
	err := utils.ExecuteCommandRun("mount", "-o", mountData.MntOpts, mountData.DevicePath, mountPath)
	// Use syscall.Mount command to mount the bricks
	if err != nil {
		return err
	}

	return nil
}

//GetVgName creates the device path for lvm snapshot
func GetVgName(mountDevice string) (string, error) {

	out, err := utils.ExecuteCommandOutput("/sbin/lvs", "--noheadings", "-o", "vg_name", mountDevice)
	if err != nil {
		return "", err
	}

	volGroup := strings.TrimSpace(string(out))
	return volGroup, nil
}

//RemoveBrickSnapshot removes an lvm of a brick
func RemoveBrickSnapshot(devicePath string) error {
	return utils.ExecuteCommandRun(RemoveCommand, "-f", devicePath)
}

//LVSnapshot takes lvm snapshot of a b
func LVSnapshot(originDevice, DevicePath string) error {

	cmd := exec.Command(CreateCommand, "-s", originDevice, "--setactivationskip", "n", "--name", DevicePath)
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the child to exit
	errStatus := cmd.Wait()
	if errStatus != nil {
		// Child exited with error
		return errStatus
	}
	return nil
}

//UpdateFsLabel sets new nabel on the device
func UpdateFsLabel(DevicePath, FsType string) error {
	uuid := uuid.NewRandom().String()
	uuid = strings.Replace(uuid, "-", "", -1)
	switch FsType {
	case "xfs":
		label := uuid[:12]
		err := utils.ExecuteCommandRun("xfs_admin", "-L", label, DevicePath)
		if err != nil {
			return err
		}
	case "ext4":
		fallthrough
	case "ext3":
		fallthrough
	case "ext2":
		label := uuid[:16]
		err := utils.ExecuteCommandRun("tune2fs", "-L", label, DevicePath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Changing file-system label of %s is not supported as of now", FsType)
	}
	return nil
}
