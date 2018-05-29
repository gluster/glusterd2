package deviceutils

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	maxMetadataSizeGb = 16
	chunkSize         = "1280k"
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

// MbToKb converts Value from Mb to Kb
func MbToKb(value uint64) uint64 {
	return value * 1024
}

// GbToKb converts Value from Gb to Kb
func GbToKb(value uint64) uint64 {
	return value * 1024 * 1024
}

// GetVgAvailableSize gets available size of given Vg
func GetVgAvailableSize(vgname string) (uint64, uint64, error) {
	out, err := exec.Command("vgdisplay", "-c", vgname).Output()
	if err != nil {
		return 0, 0, err
	}
	vgdata := strings.Split(strings.TrimRight(string(out), "\n"), ":")

	if len(vgdata) != 17 {
		return 0, 0, errors.New("Failed to get free size of VG: " + vgname)
	}

	// Physical extent size index is 12
	extentSize, err := strconv.ParseUint(vgdata[12], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	// Free Extents index is 15
	freeExtents, err := strconv.ParseUint(vgdata[15], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return extentSize * freeExtents, extentSize, nil
}

// GetPoolMetadataSize calculates the thin pool metadata size based on the given thin pool size
func GetPoolMetadataSize(poolsize uint64) uint64 {
	// https://access.redhat.com/documentation/en-us/red_hat_gluster_storage/3.3/html-single/administration_guide/#Brick_Configuration
	// Minimum metadata size required is 0.5% and Max upto 16GB

	metadataSize := uint64(float64(poolsize) * 0.005)
	if metadataSize > GbToKb(maxMetadataSizeGb) {
		metadataSize = GbToKb(maxMetadataSizeGb)
	}
	return metadataSize
}

// CreateTP creates LVM Thin Pool
func CreateTP(vgname, tpname string, tpsize, metasize uint64) error {
	// TODO: Chunksize adjust based on RAID/JBOD
	return exec.Command("lvcreate",
		"--thin", vgname+"/"+tpname,
		"--size", fmt.Sprintf("%dK", tpsize),
		"--poolmetadatasize", fmt.Sprintf("%dK", metasize),
		"-c", chunkSize,
		"--zero", "n",
	).Run()
}

// CreateLV creates LVM Logical Volume
func CreateLV(vgname, tpname, lvname string, lvsize uint64) error {
	return exec.Command("lvcreate",
		"--virtualsize", fmt.Sprintf("%dK", lvsize),
		"--thin",
		"--name", lvname,
		vgname+"/"+tpname,
	).Run()
}

// MakeXfs creates XFS filesystem
func MakeXfs(dev string) error {
	// TODO: Adjust -d su=<>,sw=<> based on RAID/JBOD
	return exec.Command("mkfs.xfs",
		"-i", "size=512",
		"-n", "size=8192",
		dev,
	).Run()
}

// BrickMount mounts the brick LV
func BrickMount(dev, mountdir string) error {
	return exec.Command("mount",
		"-o", "rw,inode64,noatime,nouuid",
		dev,
		mountdir,
	).Run()
}

// BrickUnmount unmounts the Brick
func BrickUnmount(mountdir string) error {
	return exec.Command("umount", mountdir).Run()
}

// RemoveLV removes Logical Volume
func RemoveLV(vgName, lvName string) error {
	return exec.Command("lvremove", "-f", vgName+"/"+lvName).Run()
}
