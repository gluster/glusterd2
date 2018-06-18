package brick

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/api"
	"golang.org/x/sys/unix"
)

type Brickinfo api.BrickInfo

// SizeInfo represents sizing information.
type SizeInfo struct {
	Capacity uint64
	Used     uint64
	Free     uint64
}

//Brickstatus gives status of brick
type Brickstatus struct {
	Info      Brickinfo
	Online    bool
	Pid       int
	Port      int
	FS        string
	MountOpts string
	Device    string
	Size      SizeInfo
}

func (b *Brickinfo) String() string {
	return b.PeerID.String() + ":" + b.Path
}

// StringMap returns a map[string]string representation of the Brickinfo
func (b *Brickinfo) StringMap() map[string]string {
	m := make(map[string]string)

	m["brick.id"] = b.ID.String()
	m["brick.hostname"] = b.Hostname
	m["brick.peerid"] = b.PeerID.String()
	m["brick.path"] = b.Path
	m["brick.volumename"] = b.VolumeName
	m["brick.volumeid"] = b.VolumeID.String()

	return m
}

// Validate checks if brick path is valid, if brick is a mount point,
// if brick is on root partition and if it has xattr support.
func (b *Brickinfo) Validate(check InitChecks) error {

	var (
		brickStat unix.Stat_t
		err       error
	)

	if err = validatePathLength(b.Path); err != nil {
		return err
	}

	if _, err = os.Stat(b.Path); os.IsNotExist(err) {
		if check.CreateBrickDir {
			if err = os.MkdirAll(b.Path, 0775); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err = unix.Lstat(b.Path, &brickStat); err != nil {
		return err
	}

	if (brickStat.Mode & unix.S_IFMT) != unix.S_IFDIR {
		return fmt.Errorf("Brick path %s is not a directory", b.Path)
	}

	if check.IsMount {
		if err = validateIsBrickMount(&brickStat, b.Path); err != nil {
			return err
		}
	}

	if check.IsOnRoot {
		if err = validateIsOnRootDevice(&brickStat); err != nil {
			return err
		}
	}

	if err = validateXattrSupport(b.Path); err != nil {
		return err
	}

	if check.IsInUse {
		if err = validateBrickInUse(b.Path); err != nil {
			return err
		}
	}

	return nil
}
