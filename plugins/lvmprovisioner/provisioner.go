package lvmprovisioner

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/gluster/glusterd2/pkg/fsutils"
	"github.com/gluster/glusterd2/pkg/lvmutils"
)

var mountOpts = "rw,inode64,noatime,nouuid"

// Provisioner represents lvm provisioner plugin
type Provisioner struct{}

func getVgName(devpath string) string {
	return "gluster" + strings.Replace(devpath, "/", "-", -1)
}

func getLvName(brickid string) string {
	return "lv-" + brickid
}

func getTpName(brickid string) string {
	return "tp-" + brickid
}

func getBrickDev(devpath, brickid string) string {
	return "/dev/" + getVgName(devpath) + "/" + getLvName(brickid)
}

// Register creates pv and vg for a given device
func (p Provisioner) Register(devpath string) error {
	err := lvmutils.CreatePV(devpath)
	if err != nil {
		return err
	}
	return lvmutils.CreateVG(devpath, getVgName(devpath))
}

// AvailableSize returns available size in the given device
func (p Provisioner) AvailableSize(devpath string) (uint64, uint64, error) {
	return lvmutils.GetVgAvailableSize(getVgName(devpath))
}

// Unregister removes VG and PV
func (p Provisioner) Unregister(devpath string) error {
	err := lvmutils.RemoveVG(getVgName(devpath))
	if err != nil {
		return err
	}
	return lvmutils.RemovePV(devpath)
}

// CreateBrick creates thinpool and lv for given size
func (p Provisioner) CreateBrick(devpath, brickid string, size uint64, bufferFactor float64) error {
	vgname := getVgName(devpath)
	tpsize := uint64(float64(size) * bufferFactor)
	tpname := getTpName(brickid)
	lvname := getLvName(brickid)
	metasize := lvmutils.GetTpMetadataSize(tpsize)

	err := lvmutils.CreateTP(vgname, tpname, tpsize, metasize)
	if err != nil {
		return err
	}
	return lvmutils.CreateLV(vgname, tpname, lvname, size)
}

// CreateBrickFS creates the filesystem as requested
func (p Provisioner) CreateBrickFS(devpath, brickid, fstype string) error {
	brickdev := getBrickDev(devpath, brickid)
	switch fstype {
	case "xfs":
		return fsutils.MakeXfs(brickdev)
	default:
		return errors.New("unsupported filesystem")
	}
}

// CreateBrickDir creates brick directory inside mount
func (p Provisioner) CreateBrickDir(brickPath string) error {
	return os.MkdirAll(brickPath, os.ModeDir|os.ModePerm)
}

// MountBrick mounts the brick to the given brick path
func (p Provisioner) MountBrick(devpath, brickid, brickPath string) error {
	mountdir := path.Dir(brickPath)
	brickdev := getBrickDev(devpath, brickid)
	err := os.MkdirAll(mountdir, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}
	return fsutils.Mount(brickdev, mountdir, mountOpts)
}

// UnmountBrick unmounts the brick
func (p Provisioner) UnmountBrick(brickPath string) error {
	mountdir := path.Dir(brickPath)
	return fsutils.Unmount(mountdir)
}

// RemoveBrick removes the brick LV and Thinpool
func (p Provisioner) RemoveBrick(devpath, brickid string) error {
	vgname := getVgName(devpath)
	lvname := getLvName(brickid)

	tpname, err := lvmutils.GetThinpoolName(vgname, lvname)
	if err != nil {
		return err
	}

	err = lvmutils.RemoveLV(vgname, lvname)
	if err != nil {
		return err
	}
	// Remove Thin Pool if LV count is zero, Thinpool will
	// have more LVs in case of snapshots and clones
	numLvs, err := lvmutils.NumberOfLvs(vgname, tpname)
	if err != nil {
		return err
	}

	if numLvs == 0 {
		err = lvmutils.RemoveLV(vgname, tpname)
		if err != nil {
			return err
		}
	}
	return nil
}
