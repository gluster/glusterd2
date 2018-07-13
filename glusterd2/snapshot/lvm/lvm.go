package lvm

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
)

//GetBinPath returns binary path of given name, returns null on error
func GetBinPath(name string) string {
	if str, err := exec.LookPath(name); err == nil {
		return str
	}
	return ""
}

var (
	//CreateCommand is path to lvcreate
	CreateCommand = GetBinPath("lvcreate")
	//RemoveCommand is path to lvremove
	RemoveCommand = GetBinPath("lvremove")
	//PvCreateCommand is path to pvcreate
	PvCreateCommand = GetBinPath("pvcreate")
	//VgCreateCommand is path to vgcreate
	VgCreateCommand = GetBinPath("vgcreate")
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

//IsThinLV check for lvm compatibility for a path
func IsThinLV(brickPath string) bool {
	mountRoot, err := volume.GetBrickMountRoot(brickPath)
	if err != nil {
		return false
	}

	mntInfo, err := volume.GetBrickMountInfo(mountRoot)
	if err != nil {
		return false
	}

	data, err := GetLvsData(mntInfo.FsName)
	if err != nil {
		return false
	}

	thinLV := data.PoolLV
	if thinLV == "" {
		return false
	}
	return true
}

//MountSnapshotDirectory will mount the snapshot bricks to the given path
func MountSnapshotDirectory(mountPath string, mountData brick.MountInfo) error {
	err := utils.ExecuteCommandRun("mount", "-o", mountData.MntOpts, mountData.DevicePath, mountPath)
	// Use syscall.Mount command to mount the bricks
	if err != nil {
		return err
	}

	return nil
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

//RemoveBrickSnapshot removes an lvm of a brick
func RemoveBrickSnapshot(devicePath string) error {
	return utils.ExecuteCommandRun(RemoveCommand, "-f", devicePath)
}

//LVSnapshot takes lvm snapshot of a b
func LVSnapshot(originDevice, DevicePath string) error {

	cmd := exec.Command(CreateCommand, "-s", originDevice, "--setactivationskip", "n", "--name", DevicePath)
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the child to exit
	errStatus := cmd.Wait()
	return errStatus
}

//UpdateFsLabel sets new nabel on the device
func UpdateFsLabel(DevicePath, FsType string) error {
	uuid := uuid.NewRandom().String()
	uuid = strings.Replace(uuid, "-", "", -1)
	switch FsType {
	case "xfs":
		label := uuid[:12]
		err := utils.ExecuteCommandRun("xfs_admin", "-L", label, DevicePath)
		if err != nil {
			return err
		}
	case "ext4":
		fallthrough
	case "ext3":
		fallthrough
	case "ext2":
		label := uuid[:16]
		err := utils.ExecuteCommandRun("tune2fs", "-L", label, DevicePath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Changing file-system label of %s is not supported as of now", FsType)
	}
	return nil
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
