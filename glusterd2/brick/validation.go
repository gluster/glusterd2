package brick

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/pkg/errors"
	"golang.org/x/sys/unix"
)

const (
	testXattrKey     = "trusted.glusterfs.testing-xattr-support"
	volumeIDXattrKey = "trusted.glusterfs.volume-id"
	gfidXattrKey     = "trusted.gfid"
)

// InitChecks is a set of checks to be run on a brick
type InitChecks struct {
	IsInUse  bool
	IsMount  bool
	IsOnRoot bool
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

func validateBrickInUse(brickPath string) error {
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
