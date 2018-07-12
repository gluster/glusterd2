package brick

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	"golang.org/x/sys/unix"
)

const (
	testXattrKey      = "trusted.glusterfs.testing-xattr-support"
	gfidXattrKey      = "trusted.gfid"
	volumeIDXattrKey  = "trusted.glusterfs.volume-id"
	volumeIDXattrSize = 16
)

// InitChecks is a set of checks to be run on a brick
type InitChecks struct {
	WasInUse       bool
	IsMount        bool
	IsOnRoot       bool
	CreateBrickDir bool
}

// PrepareChecks initializes InitChecks based on req
func PrepareChecks(force bool, req map[string]bool) *InitChecks {
	c := &InitChecks{}

	if force {
		// skip all checks if force is set to true

		c.CreateBrickDir = true
		return c
	}

	// do all the checks except the ones explicitly excluded
	c.WasInUse = true
	c.IsOnRoot = true
	c.IsMount = true
	c.CreateBrickDir = false

	if value, ok := req["reuse-bricks"]; ok && value {
		c.WasInUse = false
	}

	if value, ok := req["allow-root-dir"]; ok && value {
		c.IsOnRoot = false
	}

	if value, ok := req["allow-mount-as-brick"]; ok && value {
		c.IsMount = false
	}
	if value, ok := req["create-brick-dir"]; ok && value {
		c.CreateBrickDir = true
	}

	return c
}
func validatePathLength(path string) error {

	if len(filepath.Clean(path)) >= syscall.PathMax {
		return errors.ErrBrickPathTooLong
	}

	subdirs := strings.Split(path, string(os.PathSeparator))
	for _, subdir := range subdirs {
		if len(subdir) >= syscall.PathMax {
			return errors.ErrSubDirPathTooLong
		}
	}

	return nil
}

func validateIsBrickMount(brickStat *unix.Stat_t, brickPath string) error {

	var parentStat unix.Stat_t
	if err := unix.Lstat(path.Dir(brickPath), &parentStat); err != nil {
		return err
	}

	if brickStat.Dev != parentStat.Dev {
		return errors.ErrBrickIsMountPoint
	}

	return nil
}

func validateIsOnRootDevice(brickStat *unix.Stat_t) error {

	var rootStat unix.Stat_t
	if err := unix.Lstat("/", &rootStat); err != nil {
		return err
	}

	if brickStat.Dev == rootStat.Dev {
		return errors.ErrBrickUnderRootPartition
	}

	return nil
}

func validateXattrSupport(brickPath string) error {
	defer unix.Removexattr(brickPath, testXattrKey)
	return unix.Setxattr(brickPath, testXattrKey, []byte("payload"), 0)
}

// validateBrickWasUsed checks if the path was ever used a brick for a volume
// by checking if the path has certain glusterfs specific xattrs
func validateBrickWasUsed(brickPath string) error {
	keys := []string{gfidXattrKey, volumeIDXattrKey}
	for path := brickPath; path != "/"; path = filepath.Dir(path) {
		for _, key := range keys {
			size, err := unix.Getxattr(path, key, nil)
			if err != nil {
				continue
			} else if size > 0 {
				return fmt.Errorf("Xattr %s already present on %s", key, path)
			} else {
				return nil
			}
		}
	}
	return nil
}

// isBrickInActiveUse checks if the path belongs to another active brick
// belonging to an active volume currently present in this cluster.
func isBrickInActiveUse(brickPath string, allLocalBricks []Brickinfo) error {

	volumeIDBytes := make([]byte, volumeIDXattrSize)
	size, err := unix.Getxattr(brickPath, volumeIDXattrKey, volumeIDBytes)
	if err != nil || size != volumeIDXattrSize {
		// absence of xattr or error in fetching it isn't an error for
		// this purpose
		return nil
	}

	volumeID := uuid.UUID(volumeIDBytes)

	for _, b := range allLocalBricks {
		if uuid.Equal(volumeID, b.VolumeID) {
			return fmt.Errorf("Brick path %s is already in use by a volume (name=%s;id=%s)",
				brickPath, b.VolumeName, b.VolumeID)
		}
		if strings.HasPrefix(brickPath, b.Path) {
			return fmt.Errorf("Path %s is a subdirectory of another existing brick with path %s belonging to volume (name=%s;id=%s)",
				brickPath, b.Path, b.VolumeName, b.VolumeID)
		}
	}

	return nil
}
