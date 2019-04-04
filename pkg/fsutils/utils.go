package fsutils

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
)

// MakeXfs creates XFS filesystem
func MakeXfs(dev string, mkfsOpts ...string) error {
	mkfsOpts = append([]string{dev}, mkfsOpts...)
	// TODO: Adjust -d su=<>,sw=<> based on RAID/JBOD
	return utils.ExecuteCommandRun("mkfs.xfs",
		mkfsOpts...,
	)
}

//UpdateFsLabel sets new nabel on the device
func UpdateFsLabel(DevicePath, FsType string) error {
	uuid := uuid.NewRandom().String()
	uuid = strings.Replace(uuid, "-", "", -1)
	switch FsType {
	case "xfs":
		label := uuid[:12]
		if err := utils.ExecuteCommandRun("xfs_admin", "-L", label, DevicePath); err != nil {
			return err
		}
	case "ext4":
		fallthrough
	case "ext3":
		fallthrough
	case "ext2":
		label := uuid[:16]
		if err := utils.ExecuteCommandRun("tune2fs", "-L", label, DevicePath); err != nil {
			return err
		}
	default:
		return fmt.Errorf("changing file-system label of %s is not supported as of now", FsType)
	}
	return nil
}

// StatFsData represents statvfs information
type StatFsData struct {
	Total       uint64
	Free        uint64
	InodesTotal uint64
	InodesFree  uint64
}

// StatFs provides statvfs information of given path
func StatFs(pth string) (*StatFsData, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(pth, &fs)
	if err != nil {
		return nil, err
	}

	return &StatFsData{
		Total:       fs.Blocks * uint64(fs.Bsize),
		Free:        fs.Bfree * uint64(fs.Bsize),
		InodesTotal: fs.Files,
		InodesFree:  fs.Ffree,
	}, nil
}

// Mount mounts the brick
func Mount(dev, mountdir, mountOpts string) error {
	args := []string{}
	if mountOpts != "" {
		args = append(args, "-o", mountOpts)
	}
	args = append(args, dev, mountdir)

	return utils.ExecuteCommandRun("mount", args...)
}

// Unmount unmounts the Brick
func Unmount(mountdir string) error {
	return syscall.Unmount(mountdir, syscall.MNT_FORCE)
}
