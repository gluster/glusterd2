package brick

import (
	"fmt"

	"github.com/pborman/uuid"
	"golang.org/x/sys/unix"
)

// Type is the type of Brick
//go:generate stringer -type=Type
type Type uint16

const (
	// Brick represents default type of brick
	Brick Type = iota
	// Arbiter represents Arbiter brick type
	Arbiter
)

// Brickinfo is the static information about the brick
type Brickinfo struct {
	ID             uuid.UUID
	Hostname       string
	NodeID         uuid.UUID
	Path           string
	VolumeName     string
	VolumeID       uuid.UUID
	Type           Type
	Decommissioned bool
}

func (b *Brickinfo) String() string {
	return b.NodeID.String() + ":" + b.Path
}

// StringMap returns a map[string]string representation of the Brickinfo
func (b *Brickinfo) StringMap() map[string]string {
	m := make(map[string]string)

	m["brick.id"] = b.ID.String()
	m["brick.hostname"] = b.Hostname
	m["brick.nodeid"] = b.NodeID.String()
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
