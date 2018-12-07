package volume

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	volumeNameRE = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
)

// IsValidName validates Volume name
func IsValidName(name string) bool {
	return volumeNameRE.MatchString(name)
}

// GetRedundancy calculates redundancy count based on disperse count
func GetRedundancy(disperse uint) int {
	var temp, l, mask uint
	temp = disperse
	for temp = temp >> 1; temp != 0; temp = temp >> 1 {
		l = l + 1
	}
	mask = ^(1 << l)
	if red := disperse & mask; red != 0 {
		return int(red)
	}
	return 1
}

// isBrickPathAvailable validates whether the brick is consumed by other
// volume
func isBrickPathAvailable(peerID uuid.UUID, brickPath string) error {
	volumes, e := GetVolumes(context.TODO())
	if e != nil || volumes == nil {
		// In case cluster doesn't have any volumes configured yet,
		// treat this as success
		log.Debug("Failed to retrieve volumes")
		return nil
	}
	for _, v := range volumes {
		for _, b := range v.GetBricks() {
			if uuid.Equal(b.PeerID, peerID) && b.Path == brickPath {
				log.Error("Brick is already used by ", v.Name)
				return gderrors.ErrBrickPathAlreadyInUse
			}
		}
	}
	return nil
}

//CheckBricksStatus will give detailed information about brick
func CheckBricksStatus(volinfo *Volinfo) ([]brick.Brickstatus, error) {

	var brickStatuses []brick.Brickstatus
	mtabEntries, err := GetMounts()
	if err != nil {
		log.WithError(err).Error("Failed to read /etc/mtab file.")
		return brickStatuses, err
	}

	for _, binfo := range volinfo.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(binfo)
		if err != nil {
			return brickStatuses, err
		}

		s := brick.Brickstatus{
			Info: binfo,
		}

		if pidOnFile, err := daemon.ReadPidFromFile(brickDaemon.PidFile()); err == nil {
			if _, err := daemon.GetProcess(pidOnFile); err == nil {
				s.Online = true
				s.Pid = pidOnFile
				s.Port, _ = pmap.RegistrySearch(binfo.Path)
			}
		}

		var fstat syscall.Statfs_t
		if err := syscall.Statfs(binfo.Path, &fstat); err != nil {
			log.WithError(err).WithField("path",
				binfo.Path).Error("syscall.Statfs() failed")
		} else {
			s.Size = *(brick.CreateSizeInfo(&fstat))
		}

		for _, m := range mtabEntries {
			if strings.HasPrefix(binfo.Path, m.MntDir) {
				s.MountOpts = m.MntOpts
				s.Device = m.FsName
				s.FS = m.MntType
			}
		}

		brickStatuses = append(brickStatuses, s)
	}

	return brickStatuses, nil
}

//GetBrickMountRoot return root of a brick mount
func GetBrickMountRoot(brickPath string) (string, error) {
	brickStat, err := os.Stat(brickPath)
	if err != nil {
		return "", err
	}
	brickSt := brickStat.Sys().(*syscall.Stat_t)
	for dirPath := brickPath; dirPath != "/"; {
		dir := path.Dir(dirPath)
		mntStat, err := os.Stat(dir)
		if err != nil {
			return "", err
		}
		if mntSt := mntStat.Sys().(*syscall.Stat_t); brickSt.Dev != mntSt.Dev {
			return dirPath, nil
		}
		dirPath = dir
	}

	mntStat, err := os.Stat("/")
	if err != nil {
		return "", err
	}
	if mntSt := mntStat.Sys().(*syscall.Stat_t); brickSt.Dev == mntSt.Dev {
		return "/", nil
	}
	return "", errors.New("failed to get mount root")
}

//IsMountNotFoundError returns true if error matches
func IsMountNotFoundError(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "mount point not found")
	}
	return false
}

//GetBrickMountInfo return mount related information
func GetBrickMountInfo(mountRoot string) (*Mntent, error) {
	realMountRoot, err := filepath.EvalSymlinks(mountRoot)
	if err != nil {
		return nil, err
	}
	mtabEntries, err := GetMounts()
	if err != nil {
		return nil, err
	}

	for _, entry := range mtabEntries {
		if entry.MntDir == mountRoot || entry.MntDir == realMountRoot {
			return entry, nil
		}
	}
	return nil, errors.New("mount point not found")

}

//CreateSubvolInfo parses subvol information for response
func CreateSubvolInfo(sv *[]Subvol) []api.Subvol {
	var subvols []api.Subvol

	for _, subvol := range *sv {
		var blist []api.BrickInfo
		for _, b := range subvol.Bricks {
			blist = append(blist, brick.CreateBrickInfo(&b))
		}

		subvols = append(subvols, api.Subvol{
			Name:                    subvol.Name,
			Type:                    api.SubvolType(subvol.Type),
			Bricks:                  blist,
			ReplicaCount:            subvol.ReplicaCount,
			ArbiterCount:            subvol.ArbiterCount,
			DisperseCount:           subvol.DisperseCount,
			DisperseDataCount:       subvol.DisperseCount - subvol.RedundancyCount,
			DisperseRedundancyCount: subvol.RedundancyCount,
		})
	}
	return subvols
}

//CreateVolumeInfoResp parses volume information for response
func CreateVolumeInfoResp(v *Volinfo) *api.VolumeInfo {

	resp := &api.VolumeInfo{
		ID:        v.ID,
		Name:      v.Name,
		Type:      api.VolType(v.Type),
		Transport: v.Transport,
		DistCount: v.DistCount,
		Capacity:  v.Capacity,
		State:     api.VolState(v.State),
		Options:   v.Options,
		Subvols:   CreateSubvolInfo(&v.Subvols),
		Metadata:  v.Metadata,
		SnapList:  v.SnapList,
	}

	// for common use cases, replica count of the volume is usually the
	// replica count of any one of the subvols and we take replica count
	// from the first subvol
	resp.ReplicaCount = resp.Subvols[0].ReplicaCount
	resp.ArbiterCount = resp.Subvols[0].ArbiterCount
	resp.DisperseCount = resp.Subvols[0].DisperseCount
	resp.DisperseDataCount = resp.Subvols[0].DisperseDataCount
	resp.DisperseRedundancyCount = resp.Subvols[0].DisperseRedundancyCount

	return resp
}

//IsSnapshotProvisioned will return true if volume is provisioned through snapshot creation
func (v *Volinfo) IsSnapshotProvisioned() bool {
	return (v.GetProvisionType().IsSnapshotProvisioned())
}

//IsAutoProvisioned will return true if volume is automatically provisioned
func (v *Volinfo) IsAutoProvisioned() bool {
	return (v.GetProvisionType().IsAutoProvisioned())
}

//GetProvisionType will return true the type of provision state
func (v *Volinfo) GetProvisionType() brick.ProvisionType {

	var provisionType brick.ProvisionType

	provisionValue, ok := v.Metadata[brick.ProvisionKey]
	if !ok {
		provisionType = brick.ManuallyProvisioned
	} else {
		provisionType = brick.ProvisionType(provisionValue)
	}
	return provisionType
}

//CleanBricks will Unmount the bricks and delete lv, thinpool
func CleanBricks(volinfo *Volinfo) error {
	for _, b := range volinfo.GetLocalBricks() {
		// UnMount the Brick if mounted
		mountRoot := strings.TrimSuffix(b.Path, b.MountInfo.BrickDirSuffix)
		_, err := GetBrickMountInfo(mountRoot)
		if err != nil {
			if !IsMountNotFoundError(err) {
				log.WithError(err).WithField("path", mountRoot).
					Error("unable to get mount info")
				return err
			}
		} else {
			err := lvmutils.UnmountLV(mountRoot)
			if err != nil {
				log.WithError(err).WithField("path", mountRoot).
					Error("brick unmount failed")
				return err
			}
		}

		parts := strings.Split(b.MountInfo.DevicePath, "/")
		if len(parts) != 4 {
			return errors.New("unable to parse device path")
		}
		vgname := parts[2]
		lvname := parts[3]
		tpname, err := lvmutils.GetThinpoolName(vgname, lvname)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"vg-name": vgname,
				"lv-name": lvname,
			}).Error("failed to get thinpool name")
			return err
		}

		// Remove LV
		err = lvmutils.RemoveLV(vgname, lvname)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"vg-name": vgname,
				"lv-name": lvname,
			}).Error("lv remove failed")
			return err
		}

		if !deviceutils.IsVgExist(vgname) {
			continue
		}

		// Remove Thin Pool if LV count is zero, Thinpool will
		// have more LVs in case of snapshots and clones
		numLvs, err := lvmutils.NumberOfLvs(vgname, tpname)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"vg-name": vgname,
				"tp-name": tpname,
			}).Error("failed to get number of lvs")
			return err
		}

		if numLvs == 0 {
			err = lvmutils.RemoveLV(vgname, tpname)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"vg-name": vgname,
					"tp-name": tpname,
				}).Error("thinpool remove failed")
				return err
			}
		}

		// Update current Vg free size
		err = deviceutils.UpdateDeviceFreeSizeByVg(gdctx.MyUUID.String(), vgname)
		if err != nil {
			log.WithError(err).WithField("vg-name", vgname).
				Error("failed to update available size of a device")
			return err
		}
	}
	return nil
}
