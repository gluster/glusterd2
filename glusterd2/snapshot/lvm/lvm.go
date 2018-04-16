package lvm

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
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

//IsbrickComptable check for lvm comptability for a path
func IsbrickComptable(brick *brick.Brickinfo) bool {
	mountRoot, err := volume.GetBrickMountRoot(brick.Path)
	if err != nil {
		return false
	}

	mntInfo, err := volume.GetBrickMountInfo(mountRoot)
	if err != nil {
		return false
	}

	out, err := exec.Command("/sbin/lvs", "--noheadings", "-o", "pool_lv", mntInfo.FsName).Output()
	if err != nil {
		return false
	}

	thinLV := strings.TrimSpace(string(out))

	if thinLV == "" {
		return false
	}
	return true
}

//CheckBricksCompatability will verify the brickes are lvm compatable
func CheckBricksCompatability(volinfo *volume.Volinfo) []string {

	var paths []string
	for _, subvol := range volinfo.Subvols {
		for _, brick := range subvol.Bricks {
			if IsbrickComptable(&brick) != true {
				paths = append(paths, brick.String())
			}
		}
	}
	return paths
}

//MountSnapshotDirectory will mount the snapshot bricks to the given path
func MountSnapshotDirectory(mountPath string, b *brick.Brickinfo) error {
	_, err := exec.Command("mount", "-o", b.MntOpts, b.DevicePath, mountPath).Output()
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

//GetDevicePath creates the device path for lvm snapshot
func GetDevicePath(mountDevice, snapName string, brickCount int) (string, error) {

	out, err := exec.Command("/sbin/lvs", "--noheadings", "-o", "vg_name", mountDevice).Output()
	if err != nil {
		return "", err
	}

	volGroup := strings.TrimSpace(string(out))
	devicePath := "/dev/" + volGroup + "/" + snapName + "_" + strconv.Itoa(brickCount)
	return devicePath, nil
}

//RemoveBrickSnapshot removes an lvm of a brick
func RemoveBrickSnapshot(snapBrick *brick.Brickinfo) error {
	_, err := exec.Command(RemoveCommand, "f", snapBrick.DevicePath).Output()
	return err
}

//BrickSnapshot takes lvm snapshot of a brick
func BrickSnapshot(snapBrick *brick.Brickinfo, b *brick.Brickinfo) error {
	length := len(b.Path) - len(snapBrick.Mountdir)
	mountRoot := b.Path[:length]
	mntInfo, err := volume.GetBrickMountInfo(mountRoot)
	if err != nil {
		return err
	}

	cmd := exec.Command(CreateCommand, "-s", mntInfo.FsName, "--setactivationskip", "n", "--name", snapBrick.DevicePath)
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
			"devicePath":   snapBrick.DevicePath,
			"status":       errStatus,
		}).Debug("Child exited")

		if errStatus != nil {
			// Child exited with error
			return errStatus
		}
		err = updateFsLabel(snapBrick, b)

	}
	return err
}

func updateFsLabel(snapBrick, b *brick.Brickinfo) error {
	uuid := uuid.NewRandom().String()
	uuid = strings.Replace(uuid, "-", "", -1)
	switch snapBrick.FsType {
	case "xfs":
		label := uuid[:12]
		_, err := exec.Command("xfs_admin", "-L", label, snapBrick.DevicePath).Output()
		if err != nil {
			return err
		}
	case "ext4":
		fallthrough
	case "ext3":
		fallthrough
	case "ext2":
		label := uuid[:16]
		_, err := exec.Command("tune2fs", "-L", label, b.DevicePath).Output()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Changing file-system label of %s is not supported as of now", b.FsType)
	}
	return nil
}
