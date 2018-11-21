package lvmtest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/lvmutils"
)

const (
	lvmPrefix    string = "patchy_snap"
	devicePrefix string = "/dev/gluster_loop"
)

var (
	xfsFormat    = lvmutils.GetBinPath("mkfs.xfs")
	fallocateBin = lvmutils.GetBinPath("fallocate")
	mknodBin     = lvmutils.GetBinPath("mknod")
	brickPrefix  string
)

func verifyLVM() bool {
	out, err := exec.Command(lvmutils.CreateCommand, "--help").Output()
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
		if err := os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
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
		devicePath := devicePrefix + strconv.Itoa(i)

		vg := fmt.Sprintf("%s_vg_%d", lvmPrefix, i)
		poolPath := fmt.Sprintf("/dev/%s/thinpool", vg)
		xfsPath := fmt.Sprintf("/dev/%s/brick_lvm", vg)

		if err := exec.Command(lvmutils.PvCreateCommand, devicePath).Run(); err != nil {
			return err
		}

		if err := exec.Command(lvmutils.VgCreateCommand, vg, devicePath).Run(); err != nil {
			return err
		}

		if err := exec.Command(lvmutils.CreateCommand, "-L", thinpoolSize, "-T", poolPath).Run(); err != nil {
			return err
		}

		if err := exec.Command(lvmutils.CreateCommand, "-V", virtualSize, "-T", poolPath, "-n", "brick_lvm").Run(); err != nil {
			return err
		}

		if err := exec.Command(xfsFormat, "-f", xfsPath).Run(); err != nil {
			return err
		}
		if err := exec.Command("mount", "-t", "xfs", "-o", "nouuid", xfsPath, brickPath).Run(); err != nil {
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

		if err := syscall.Unmount(brickPath, syscall.MNT_FORCE); err != nil && !force {
			return err
		}
		if err := os.RemoveAll(brickPath); err != nil && !force {
			return err
		}
		if err := exec.Command(lvmutils.RemoveCommand, "-f", vg).Run(); err != nil && !force {
			return err
		}
		if err := exec.Command(lvmutils.VgRemoveCommand, "-f", vg).Run(); err != nil && !force {
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
		devicePath := devicePrefix + strconv.Itoa(i)

		if err := exec.Command(lvmutils.PvRemoveCommand, "-f", devicePath).Run(); err != nil && !force {
			return err
		}
		if err := exec.Command("losetup", "-d", devicePath).Run(); err != nil && !force {
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
		devicePath := devicePrefix + strconv.Itoa(i)
		//TODO replace exec command with syscall.Fallocate
		if err := exec.Command(fallocateBin, "-l", size, vhdPath).Run(); err != nil {
			return err
		}
		if err := exec.Command(mknodBin, devicePath, "b", "7", strconv.Itoa(i)).Run(); err != nil {
			return err
		}
		loosetupCmd := exec.Command("losetup", devicePath, vhdPath)
		if err := loosetupCmd.Run(); err != nil {
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
	if err = createVHD(num, "300M"); err != nil {
		return brickPath, err
	}

	if err = createLV(num, "200M", "150M"); err != nil {
		return brickPath, err
	}
	return brickPath, nil

}

//Cleanup will kill all process, and remove mount points
func Cleanup(baseWorkdir, prefix string, brickCount int) {

	brickPrefix = prefix
	exec.Command("pkill", "gluster").Run()
	time.Sleep(3 * time.Second)

	mtabEntries, err := volume.GetMounts()
	if err != nil {
		return
	}
	for _, m := range mtabEntries {
		if strings.HasPrefix(m.MntDir, baseWorkdir) {

			//Remove any dangling mount pounts
			syscall.Unmount(m.MntDir, syscall.MNT_FORCE|syscall.MNT_DETACH)
		}
	}

	deleteLV(brickCount, true)
	deleteVHD(brickCount, true)

	vg := fmt.Sprintf("%s_vg_", lvmPrefix)
	out, err := exec.Command(lvmutils.LVSCommand, "--noheadings", "-o", "vg_name").Output()
	if err != nil {
		// TODO: log failure here
		return
	}
	for _, entry := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(entry, vg) {
			exec.Command(lvmutils.RemoveCommand, "-f", entry)
			exec.Command(lvmutils.VgRemoveCommand, "-f", entry)
		}
	}
	os.RemoveAll(brickPrefix)

}

//CleanupLvmBricks provides an lvm mount point created using loop back devices
func CleanupLvmBricks(prefix string, num int) error {
	var err error
	brickPrefix = prefix
	if !verifyLVM() {
		return errors.New("lvm or thinlv is not available on the machine")
	}

	if err = deleteLV(num, false); err != nil {
		return err
	}

	if err = deleteVHD(num, false); err != nil {
		return err
	}

	err = os.RemoveAll(brickPrefix)
	return err

}
