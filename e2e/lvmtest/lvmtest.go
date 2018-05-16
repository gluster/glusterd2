package lvmtest

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	brickPrefix = "/d/backends/"
	lvmPrefix   = "patchy_snap"
	pathPrefix  []string
)

func verifyLVM() bool {
	out, err := exec.Command("/sbin/lvcreate", "--help").Output()
	if err != nil {
		return false
	}
	thinLV := strings.Contains(string(out), "thin")
	return thinLV
}

func createBrickPath(num int) ([]string, error) {

	var brickPath []string

	for i := 1; i <= num; i++ {
		prefix := brickPrefix + strconv.Itoa(i) + "/" + lvmPrefix
		pathPrefix = append(pathPrefix, prefix)
		path := prefix + "_mnt"
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
		brickPath := pathPrefix[i-1] + "_mnt"
		devicePath := pathPrefix[i-1] + "_loop"
		vg := lvmPrefix + "_vg_" + strconv.Itoa(i)
		poolPath := "/dev/" + vg + "/thinpool"
		xfsPath := "/dev/" + vg + "/brick_lvm"

		if _, err := exec.Command("/sbin/pvcreate", devicePath).Output(); err != nil {
			return err
		}

		if _, err := exec.Command("/sbin/vgcreate", vg, devicePath).Output(); err != nil {
			return err
		}

		if _, err := exec.Command("/sbin/lvcreate", "-L", thinpoolSize, "-T", poolPath).Output(); err != nil {
			return err
		}

		if _, err := exec.Command("/sbin/lvcreate", "-V", virtualSize, "-T", poolPath, "-n", "brick_lvm").Output(); err != nil {
			return err
		}

		if _, err := exec.Command("/usr/sbin/mkfs.xfs", "-f", xfsPath).Output(); err != nil {
			return err
		}
		if _, err := exec.Command("mount", "-t", "xfs", "-o", "nouuid", xfsPath, brickPath).Output(); err != nil {
			return err
		}

	}
	return nil
}

func deleteLV(num int) error {
	for i := 1; i <= num; i++ {
		brickPath := pathPrefix[i-1] + "_mnt"
		vg := lvmPrefix + "_vg_" + strconv.Itoa(i)
		if _, err := exec.Command("umount", "-f", brickPath).Output(); err != nil {
			return err
		}
		if err := os.RemoveAll(brickPath); err != nil {
			return err
		}
		if _, err := exec.Command("/sbin/vgremove", "-f", vg).Output(); err != nil {
			return err
		}

	}
	return nil

}

//create given number of virtual hard disk
func deleteVHD(num int) error {

	for i := 1; i <= num; i++ {
		vhdPath := pathPrefix[i-1] + "_vhd"
		devicePath := pathPrefix[i-1] + "_loop"
		_, err := exec.Command("losetup", "-d", devicePath).Output()
		if err != nil {
			return err
		}
		if err := os.RemoveAll(vhdPath); err != nil {
			return err
		}
		if err := os.RemoveAll(devicePath); err != nil {
			return err
		}

	}
	return nil
}

//create given number of virtual hard disk
func createVHD(num int, size string) error {

	for i := 1; i <= num; i++ {
		vhdPath := pathPrefix[i-1] + "_vhd"
		devicePath := pathPrefix[i-1] + "_loop"
		//TODO replace exec command with syscall.Fallocate
		_, err := exec.Command("/usr/bin/fallocate", "-l", size, vhdPath).Output()
		if err != nil {
			return err
		}
		_, err = exec.Command("/usr/bin/mknod", "-m", "660", devicePath, "b", "7", strconv.Itoa(i+8)).Output()
		loosetupCmd := exec.Command("losetup", devicePath, vhdPath)
		_, err = loosetupCmd.Output()
		if err != nil {
			return err

		}

	}
	return nil
}

//CreateLvmBricks provides an lvm mount point created using loop back devices
func CreateLvmBricks(num int) ([]string, error) {
	var brickPath []string
	var err error
	if lvm := verifyLVM(); lvm == false {
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

//CleanupLvmBricks provides an lvm mount point created using loop back devices
func CleanupLvmBricks(num int) error {
	var err error
	if lvm := verifyLVM(); lvm == false {
		return errors.New("lvm or thinlv is not available on the machine")
	}

	err = deleteVHD(num)
	if err != nil {
		return err
	}

	err = deleteLV(num)
	if err != nil {
		return err
	}

	err = os.RemoveAll(brickPrefix)
	return err

}
