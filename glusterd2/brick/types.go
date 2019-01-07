package brick

import (
	"fmt"
	"os"

	"github.com/pborman/uuid"
	"golang.org/x/sys/unix"
)

// Type is the type of Brick
//go:generate stringer -type=Type
type Type uint16

//ProvisionType is the way a brick is provisioned
type ProvisionType string

const (
	// Brick represents default type of brick
	Brick Type = iota
	// Arbiter represents Arbiter brick type
	Arbiter
	// ThinArbiter represents thin arbiter brick type
	ThinArbiter
)

//MountInfo is used to store mount related information of a volume
type MountInfo struct {
	BrickDirSuffix string
	DevicePath     string
	FsType         string
	MntOpts        string
}

// DeviceInfo is used to store brick device information
type DeviceInfo struct {
	TpName     string
	LvName     string
	VgName     string
	RootDevice string
}

// Brickinfo is the static information about the brick
type Brickinfo struct {
	ID             uuid.UUID
	Hostname       string
	PeerID         uuid.UUID
	Path           string
	VolumeName     string
	VolfileID      string
	VolumeID       uuid.UUID
	Type           Type
	Decommissioned bool
	PType          ProvisionType
	MountInfo
	DeviceInfo
}

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

const (
	//ProvisionKey is used to set the type of provisioning in volume metadata
	ProvisionKey string = "_brick-provision-type"

	//ManuallyProvisioned bricks will be provisioned by a user
	ManuallyProvisioned ProvisionType = ""
	//AutoProvisioned bricks will be provsioned by device manager from glusterd2
	AutoProvisioned ProvisionType = "auto"
	//SnapshotProvisioned bricks will be created by performing a snapshot of a gluster brick
	SnapshotProvisioned ProvisionType = "snapshot"
)

//IsManuallyProvisioned will return true if manually provisioned
func (p ProvisionType) IsManuallyProvisioned() bool {
	return p == ManuallyProvisioned
}

//IsAutoProvisioned will return true if auto provisioned
func (p ProvisionType) IsAutoProvisioned() bool {
	return p == AutoProvisioned
}

//IsSnapshotProvisioned will return true if provisioned via snapshot operation
func (p ProvisionType) IsSnapshotProvisioned() bool {
	return p == SnapshotProvisioned
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
func (b *Brickinfo) Validate(check InitChecks, allLocalBricks []Brickinfo) error {

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

	if check.WasInUse {
		if err = validateBrickWasUsed(b.Path); err != nil {
			return err
		}
	}

	// mandatory check that cannot be skipped forcefully
	return isBrickInActiveUse(b.Path, allLocalBricks)
}

//BrickTypeToString converts BrickType to corresponding string
func (b *Brickinfo) BrickTypeToString() string {
	switch b.Type {
	case Arbiter:
		return "arbiter"
	default:
		return "brick"
	}
}
