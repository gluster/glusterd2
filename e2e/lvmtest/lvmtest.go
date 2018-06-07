package lvmtest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gluster/glusterd2/glusterd2/snapshot/lvm"
	config "github.com/spf13/viper"
)

const (
	lvmPrefix    string = "patchy_snap"
	xfsFormat    string = "/usr/sbin/mkfs.xfs"
	fallocateBin string = "/usr/bin/fallocate"
	mknodBin     string = "/usr/bin/mknod"
)

var (
	brickPrefix string
)

func verifyLVM() bool {
	out, err := exec.Command(lvm.CreateCommand, "--help").Output()
	if err != nil {
		return false
	}
	thinLV := strings.Contains(string(out), "thin")
	return thinLV
}

func createBrickPath(num int) ([]string, error) {

	var brickPath []string

	for i := 1; i <= num; i++ {
		prefix := fmt.Sprintf("%s%d/%s", brickPrefix, i, lvmPrefix)
		path := fmt.Sprintf("%s_mnt", prefix)
		err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
		if err != nil {
			return brickPath, err
		}
		brickPath = append(brickPath, path)
	}

	return brickPath, nil
}

func createLV(num int, thinpoolSize, virtualSize string) error {

	for i := 1; i <= num; i++ {
		prefix := fmt.Sprintf("%s%d/%s", brickPrefix, i, lvmPrefix)
		brickPath := fmt.Sprintf("%s_mnt", prefix)
		devicePath := fmt.Sprintf("%s_loop", prefix)

		vg := fmt.Sprintf("%s_vg_%d", lvmPrefix, i)
		poolPath := fmt.Sprintf("/dev/%s/thinpool", vg)
		xfsPath := fmt.Sprintf("/dev/%s/brick_lvm", vg)

		if _, err := exec.Command(lvm.PvCreateCommand, devicePath).Output(); err != nil {
			return err
		}

		if _, err := exec.Command(lvm.VgCreateCommand, vg, devicePath).Output(); err != nil {
			return err
		}

		if _, err := exec.Command(lvm.CreateCommand, "-L", thinpoolSize, "-T", poolPath).Output(); err != nil {
			return err
		}

		if _, err := exec.Command(lvm.CreateCommand, "-V", virtualSize, "-T", poolPath, "-n", "brick_lvm").Output(); err != nil {
			return err
		}

		if _, err := exec.Command(xfsFormat, "-f", xfsPath).Output(); err != nil {
			return err
		}
		if _, err := exec.Command("mount", "-t", "xfs", "-o", "nouuid", xfsPath, brickPath).Output(); err != nil {
			return err
		}

	}
	return nil
}

func deleteLV(num int, force bool) error {
	for i := 1; i <= num; i++ {
		prefix := fmt.Sprintf("%s%d/%s", brickPrefix, i, lvmPrefix)
		brickPath := fmt.Sprintf("%s_mnt", prefix)
		vg := fmt.Sprintf("%s_vg_%d", lvmPrefix, i)

		if _, err := exec.Command("umount", "-f", brickPath).Output(); err != nil && !force {
			return err
		}
		if err := os.RemoveAll(brickPath); err != nil && !force {
			return err
		}
		if _, err := exec.Command(lvm.RemoveCommand, "-f", vg).Output(); err != nil && !force {
			return err
		}

	}
	return nil

}

//create given number of virtual hard disk
func deleteVHD(num int, force bool) error {

	for i := 1; i <= num; i++ {
		prefix := fmt.Sprintf("%s%d/%s", brickPrefix, i, lvmPrefix)
		vhdPath := fmt.Sprintf("%s_vhd", prefix)
		devicePath := fmt.Sprintf("%s_loop", prefix)
		_, err := exec.Command("losetup", "-d", devicePath).Output()
		if err != nil && !force {
			return err
		}
		if err := os.RemoveAll(vhdPath); err != nil && !force {
			return err
		}
		if err := os.RemoveAll(devicePath); err != nil && !force {
			return err
		}

	}
	return nil
}

//create given number of virtual hard disk
func createVHD(num int, size string) error {

	for i := 1; i <= num; i++ {
		prefix := fmt.Sprintf("%s%d/%s", brickPrefix, i, lvmPrefix)
		vhdPath := fmt.Sprintf("%s_vhd", prefix)
		devicePath := fmt.Sprintf("%s_loop", prefix)
		//TODO replace exec command with syscall.Fallocate
		_, err := exec.Command(fallocateBin, "-l", size, vhdPath).Output()
		if err != nil {
			return err
		}
		_, err = exec.Command(mknodBin, "-m", "660", devicePath, "b", "7", strconv.Itoa(i+8)).Output()
		loosetupCmd := exec.Command("losetup", devicePath, vhdPath)
		_, err = loosetupCmd.Output()
		if err != nil {
			return err

		}

	}
	return nil
}

//CreateLvmBricks provides an lvm mount point created using loop back devices
func CreateLvmBricks(prefix string, num int) ([]string, error) {
	var brickPath []string
	brickPrefix = prefix
	var err error
	if !verifyLVM() {
		return brickPath, errors.New("lvm or thinlv is not available on the machine")
	}

	brickPath, err = createBrickPath(num)
	if err != nil {
		return brickPath, err
	}
	err = createVHD(num, "300M")
	if err != nil {
		return brickPath, err
	}

	err = createLV(num, "200M", "150M")
	if err != nil {
		return brickPath, err
	}
	return brickPath, nil

}

//Cleanup will kill all process, and remove mount points
func Cleanup(prefix string, brickCount int) {

	brickPrefix = prefix
	exec.Command("pkill", "gluster").Output()
	time.Sleep(2 * time.Second)

	snapDirPrefix := config.GetString("rundir") + "/snaps/*"
	//Remove any dangling snapshot mount pounts
	exec.Command("umount", snapDirPrefix).Output()
	deleteVHD(brickCount, true)
	deleteLV(brickCount, true)
	os.RemoveAll(brickPrefix)

}

//CleanupLvmBricks provides an lvm mount point created using loop back devices
func CleanupLvmBricks(prefix string, num int) error {
	var err error
	brickPrefix = prefix
	if !verifyLVM() {
		return errors.New("lvm or thinlv is not available on the machine")
	}
	err = deleteVHD(num, false)
	if err != nil {
		return err
	}

	err = deleteLV(num, false)
	if err != nil {
		return err
	}

	err = os.RemoveAll(brickPrefix)
	return err

}
