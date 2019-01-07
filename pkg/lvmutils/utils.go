package lvmutils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
)

//TODO make it configurable with config value

//MaxSizePercentage , above this value snapshot creation won't be allowed
const MaxSizePercentage = 90.0

var (
	//CreateCommand is path to lvcreate
	CreateCommand = GetBinPath("lvcreate")
	//RemoveCommand is path to lvremove
	RemoveCommand = GetBinPath("lvremove")
	//PvCreateCommand is path to pvcreate
	PvCreateCommand = GetBinPath("pvcreate")
	//VgCreateCommand is path to vgcreate
	VgCreateCommand = GetBinPath("vgcreate")
	//VgRemoveCommand is path to vgremove
	VgRemoveCommand = GetBinPath("vgremove")
	//PvRemoveCommand is path to pvremove
	PvRemoveCommand = GetBinPath("pvremove")

	//LVSCommand is path to lvs
	LVSCommand = GetBinPath("lvs")
)

//LvsData provides the information about an thinLV
type LvsData struct {
	VgName         string
	DataPercentage float32
	LvSize         string
	PoolLV         string
}

const (
	maxMetadataSize = 16 * utils.GiB
	chunkSize       = "1280k"
)

//CreatePV is used to create physical volume.
func CreatePV(device string) error {
	return utils.ExecuteCommandRun("pvcreate", "--metadatasize=128M", "--dataalignment=256K", device)
}

//CreateVG is used to create volume group
func CreateVG(device string, vgName string) error {
	return utils.ExecuteCommandRun("vgcreate", vgName, device)
}

//RemoveVG is used to remove volume group.
func RemoveVG(vgName string) error {
	return utils.ExecuteCommandRun("vgremove", vgName)
}

//RemovePV is used to remove physical volume
func RemovePV(device string) error {
	return utils.ExecuteCommandRun("pvremove", device)
}

// GetVgAvailableSize gets available size of given Vg
func GetVgAvailableSize(vgname string) (uint64, uint64, error) {
	out, err := exec.Command("vgdisplay", "-c", "--readonly", vgname).Output()
	if err != nil {
		return 0, 0, err
	}
	vgdata := strings.Split(strings.TrimRight(string(out), "\n"), ":")

	if len(vgdata) != 17 {
		return 0, 0, errors.New("failed to get free size of VG: " + vgname)
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
	extentSize = extentSize * utils.KiB

	return extentSize * freeExtents, extentSize, nil
}

// GetPoolMetadataSize calculates the thin pool metadata size based on the given thin pool size
func GetPoolMetadataSize(poolsize uint64) uint64 {
	// https://access.redhat.com/documentation/en-us/red_hat_gluster_storage/3.3/html-single/administration_guide/#Brick_Configuration
	// Minimum metadata size required is 0.5% and Max upto 16GB ~ 17179869184 Bytes

	metadataSize := uint64(float64(poolsize) * 0.005)

	if metadataSize > maxMetadataSize {
		metadataSize = maxMetadataSize
	}

	return NormalizeSize(metadataSize)
}

// CreateTP creates LVM Thin Pool
func CreateTP(vgname, tpname string, tpsize, metasize uint64) error {
	// TODO: Chunksize adjust based on RAID/JBOD
	return utils.ExecuteCommandRun("lvcreate",
		"--thin", vgname+"/"+tpname,
		"--size", fmt.Sprintf("%dB", tpsize),
		"--poolmetadatasize", fmt.Sprintf("%dB", metasize),
		"-c", chunkSize,
		"--zero", "n",
	)
}

// CreateLV creates LVM Logical Volume
func CreateLV(vgname, tpname, lvname string, lvsize uint64) error {
	return utils.ExecuteCommandRun("lvcreate",
		"--virtualsize", fmt.Sprintf("%dB", lvsize),
		"--thin",
		"--name", lvname,
		vgname+"/"+tpname,
	)
}

// MountLV mounts the brick LV
func MountLV(dev, mountdir, mountOpts string) error {
	args := []string{}
	if mountOpts != "" {
		args = append(args, "-o", mountOpts)
	}
	args = append(args, dev, mountdir)

	return utils.ExecuteCommandRun("mount", args...)
}

// UnmountLV unmounts the Brick
func UnmountLV(mountdir string) error {
	return syscall.Unmount(mountdir, syscall.MNT_FORCE)
}

// IsDependentLvsError returns true if the error related to dependent Lvs exists
func IsDependentLvsError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "dependent volume(s). Proceed? [y/n]")
}

// IsLvNotFoundError returns true if the error is related to non existent LV error
func IsLvNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Failed to find logical volume")
}

// DeactivateLV deactivates a Logical Volume
func DeactivateLV(vgName, lvName string) error {
	return utils.ExecuteCommandRun("lvchange", "-a", "n", vgName+"/"+lvName)
}

// ActivateLV activates a Logical Volume
func ActivateLV(vgName, lvName string) error {
	return utils.ExecuteCommandRun("lvchange", "-a", "y", vgName+"/"+lvName)
}

// RemoveLV removes Logical Volume
func RemoveLV(vgName, lvName string, force bool) error {
	args := []string{"--autobackup", "y", vgName + "/" + lvName}
	if force {
		args = append(args, "-f")
	}
	return utils.ExecuteCommandRun("lvremove", args...)
}

// NumberOfLvs returns number of Lvs present in thinpool
func NumberOfLvs(vgname, tpname string) (int, error) {
	nlv := 0
	out, err := utils.ExecuteCommandOutput(
		"lvs", "--no-headings", "--readonly", "--select",
		fmt.Sprintf("vg_name=%s&&pool_lv=%s", vgname, tpname),
	)

	if err == nil {
		out := strings.Trim(string(out), " \n")
		if out == "" {
			nlv = 0
		} else {
			nlv = len(strings.Split(out, "\n"))
		}
	}
	return nlv, err
}

// GetThinpoolName gets thinpool name for a given LV
func GetThinpoolName(vgname, lvname string) (string, error) {
	out, err := utils.ExecuteCommandOutput(
		"lvs", "--no-headings", "--readonly", "--select",
		fmt.Sprintf("vg_name=%s&&lv_name=%s", vgname, lvname),
		"-o", "pool_lv",
	)
	if err == nil {
		return strings.Trim(string(out), " \n"), nil
	}

	return "", err
}

//GetBinPath returns binary path of given name, returns null on error
func GetBinPath(name string) string {
	if str, err := exec.LookPath(name); err == nil {
		return str
	}
	return ""
}

// CommonPrevalidation checks the lvm related validation for snapshot
func CommonPrevalidation(lvmCommand string) error {
	fileInfo, err := os.Stat(lvmCommand)
	if err != nil {
		return err
	}
	switch mode := fileInfo.Mode(); {

	case mode.IsRegular() == false:
		return err
	case mode&0111 == 0:
		return err
	}
	return nil
}

//FsCompatibleCheck check for lvm compatibility for a path
func FsCompatibleCheck(fsName string) bool {
	data, err := GetLvsData(fsName)
	if err != nil {
		return false
	}

	thinLV := data.PoolLV
	if thinLV == "" {
		return false
	}
	return true
}

//SizeCompatibleCheck check for lvm compatibility for a path
func SizeCompatibleCheck(fsName string) bool {
	data, err := GetLvsData(fsName)
	if err != nil {
		return false
	}

	thinPool := fmt.Sprintf("/dev/%s/%s", data.VgName, data.PoolLV)
	thinData, err := GetLvsData(thinPool)
	if err != nil {
		return false
	}
	if thinData.DataPercentage >= float32(MaxSizePercentage) {
		return false
	}

	return true
}

//GetVgName creates the device path for lvm snapshot
func GetVgName(mountDevice string) (string, error) {

	data, err := GetLvsData(mountDevice)
	if err != nil {
		return "", err
	}

	volGroup := data.VgName
	return volGroup, nil
}

//RemoveLVSnapshot removes an lvm of a brick
func RemoveLVSnapshot(devicePath string) error {
	return utils.ExecuteCommandRun(RemoveCommand, "-f", devicePath)
}

//LVSnapshot takes lvm snapshot of a b
func LVSnapshot(originDevice, DevicePath string) error {

	cmd := exec.Command(CreateCommand, "-s", originDevice, "--setactivationskip", "n", "--name", DevicePath)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for the child to exit
	errStatus := cmd.Wait()
	return errStatus
}

//CreateLvsResp creates corresponding response strcture for LvsData
func CreateLvsResp(lvs LvsData) api.LvsData {
	s := api.LvsData{
		VgName:         lvs.VgName,
		DataPercentage: lvs.DataPercentage,
		LvSize:         lvs.LvSize,
		PoolLV:         lvs.PoolLV,
	}
	return s
}

//GetLvsData creates the device path for lvm snapshot
func GetLvsData(mountDevice string) (LvsData, error) {

	out, err := exec.Command(LVSCommand, "--noheadings", "-o", "vg_name,data_percent,lv_size,pool_lv", "--separator", ":", mountDevice).Output()
	if err != nil {
		return LvsData{}, err
	}
	data := strings.Split(string(out), ":")
	dataPercentage, err := strconv.ParseFloat(data[1], 32)
	if err != nil {
		return LvsData{}, err
	}
	result := LvsData{
		VgName:         strings.TrimSpace(data[0]),
		DataPercentage: float32(dataPercentage),
		LvSize:         strings.TrimSpace(data[2]),
		PoolLV:         strings.TrimSpace(data[3]),
	}
	return result, nil
}

//CreateDevicePath creates device path for new snapshot
func CreateDevicePath(originDevice, prefix string) (string, error) {
	vG, err := GetVgName(originDevice)
	if err != nil {
		return "", err
	}
	devicePath := fmt.Sprintf("/dev/%s/%s", vG, prefix)
	if _, err = GetLvsData(devicePath); err == nil {
		//ThinLV already exist
		errMSG := fmt.Sprintf("Failed to creaite device name %s for device %s. A thinLV with same name exist.", devicePath, originDevice)
		return "", errors.New(errMSG)
	}
	return devicePath, nil
}

// ExtendLV extends the lv by the size specified, used for intelligent volume expand
func ExtendLV(totalExpansionSizePerBrick uint64, vgName string, lvName string) error {
	err := utils.ExecuteCommandRun("lvresize", "--resizefs", "--size", fmt.Sprintf("+%dB", totalExpansionSizePerBrick), fmt.Sprintf("/dev/%s/%s", vgName, lvName))
	return err
}

// ExtendMetadataPool extends the metadata pool by the size specified, used for intelligent volume expand
func ExtendMetadataPool(expansionMetadataSizePerBrick uint64, vgName string, tpName string) error {
	if expansionMetadataSizePerBrick < 1 {
		expansionMetadataSizePerBrick = 512
	}
	err := utils.ExecuteCommandRun("lvextend", "--poolmetadatasize", fmt.Sprintf("+%dB", expansionMetadataSizePerBrick), fmt.Sprintf("/dev/%s/%s", vgName, tpName))
	return err
}

// ExtendThinpool extends the thinpool by the size specified, used for intelligent volume expand
func ExtendThinpool(expansionTpSizePerBrick uint64, vgName string, tpName string) error {
	err := utils.ExecuteCommandRun("lvextend", fmt.Sprintf("-L+%dB", expansionTpSizePerBrick), fmt.Sprintf("/dev/%s/%s", vgName, tpName))
	return err
}

// NormalizeSize converts the value to multiples of 512
func NormalizeSize(size uint64) uint64 {
	rem := size % 512
	if rem > 0 {
		size = size - rem
	}
	return size
}
