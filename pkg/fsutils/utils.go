package fsutils

import (
	"syscall"

	"github.com/gluster/glusterd2/pkg/utils"
)

// MakeXfs creates XFS filesystem
func MakeXfs(dev string) error {
	// TODO: Adjust -d su=<>,sw=<> based on RAID/JBOD
	return utils.ExecuteCommandRun("mkfs.xfs",
		"-i", "size=512",
		"-n", "size=8192",
		dev,
	)
}

// Mount mounts the brick LV
func Mount(dev, mountdir, options string) error {
	return utils.ExecuteCommandRun("mount",
		"-o", options,
		dev,
		mountdir,
	)
}

// Unmount unmounts the Brick
func Unmount(mountdir string) error {
	return syscall.Unmount(mountdir, syscall.MNT_FORCE)
}
