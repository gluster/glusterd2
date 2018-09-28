package provisioners

import (
	"errors"

	"github.com/gluster/glusterd2/plugins/lvmprovisioner"
)

// Provisioner represents bricks provisioner
type Provisioner interface {
	// Register will be called when user registers a device or directory.
	// PV and VG will be created in case of lvm plugin
	Register(devpath string) error
	// AvailableSize returns available size of the device
	AvailableSize(devpath string) (uint64, uint64, error)
	// Unregister will be called when device/directory needs to be removed
	Unregister(devpath string) error
	// CreateBrick creates the brick volume
	CreateBrick(devpath, brickid string, size uint64, bufferFactor float64) error
	// CreateBrickFs will create the brick filesystem
	CreateBrickFS(devpath, brickid, fstype string) error
	// CreateBrickDir will create the brick directory
	CreateBrickDir(brickPath string) error
	// MountBrick will mount the brick
	MountBrick(devpath, brickid, brickPath string) error
	// UnmountBrick will unmount the brick
	UnmountBrick(brickPath string) error
	// RemoveBrick will remove the brick
	RemoveBrick(devpath, brickid string) error
}

const defaultProvisionerName = "lvm"

var provisionersMap map[string]Provisioner

// Get returns requested provisioner if exists
func Get(name string) (Provisioner, error) {
	prov, exists := provisionersMap[name]
	if !exists {
		return nil, errors.New("unsupported provisioner")
	}
	return prov, nil
}

// GetDefault returns default Provisioner
func GetDefault() Provisioner {
	return provisionersMap[defaultProvisionerName]
}

func init() {
	provisionersMap = map[string]Provisioner{
		"lvm": lvmprovisioner.Provisioner{},
	}
}
