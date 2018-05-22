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
	"github.com/sirupsen/logrus"
)

var (
	//CreateCommand is path to lvm create
	CreateCommand = "/sbin/lvcreate"
	//RemoveCommand is path to lvm create
	RemoveCommand = "/sbin/lvremove"
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
	/*
		logrus.WithFields(logrus.Fields{
			"device path": b.DevicePath,
			"mount path":  mountPath,
			"fs type":     b.FsType,
			"options":     b.MntOpts,
		}).Debug("Mounting the device")

		TODO use system mount command to mount the brick
		err := syscall.Mount(b.DevicePath, mountPath, b.FsType,, syscall.MS_MGC_VAL, b.MntOpts)
	*/
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
func RemoveBrickSnapshot(mountData brick.MountInfo) error {
	return utils.ExecuteCommandRun(RemoveCommand, "f", mountData.DevicePath)
}

//BrickSnapshot takes lvm snapshot of a brick
func BrickSnapshot(mountData brick.MountInfo, path string) error {
	length := len(path) - len(mountData.Mountdir)
	mountRoot := path[:length]
	mntInfo, err := volume.GetBrickMountInfo(mountRoot)
	if err != nil {
		return err
	}

	cmd := exec.Command(CreateCommand, "-s", mntInfo.FsName, "--setactivationskip", "n", "--name", mountData.DevicePath)
	err = cmd.Start()
	if err != nil {
		return err
	}

	if true {
		// Wait for the child to exit
		errStatus := cmd.Wait()
		logrus.WithFields(logrus.Fields{
			"pid":          cmd.Process.Pid,
			"mount device": mntInfo.FsName,
			"devicePath":   mountData.DevicePath,
			"status":       errStatus,
		}).Debug("Child exited")

		if errStatus != nil {
			// Child exited with error
			return errStatus
		}
		err = UpdateFsLabel(mountData.DevicePath, mountData.FsType)

	}
	return err
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
