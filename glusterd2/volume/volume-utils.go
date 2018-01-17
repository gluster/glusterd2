package volume

import (
	"errors"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// isBrickPathAvailable validates whether the brick is consumed by other
// volume
func isBrickPathAvailable(nodeID uuid.UUID, brickPath string) error {
	volumes, e := GetVolumes()
	if e != nil || volumes == nil {
		// In case cluster doesn't have any volumes configured yet,
		// treat this as success
		log.Debug("Failed to retrieve volumes")
		return nil
	}
	for _, v := range volumes {
		for _, b := range v.GetBricks() {
			if uuid.Equal(b.NodeID, nodeID) && b.Path == brickPath {
				log.Error("Brick is already used by ", v.Name)
				return gderrors.ErrBrickPathAlreadyInUse
			}
		}
	}
	return nil
}

// IsBitrotEnabled returns true if bitrot is enabled for a volume and false otherwise
func IsBitrotEnabled(v *Volinfo) bool {
	val, exists := v.Options[VkeyFeaturesBitrot]
	if exists && val == "on" {
		return true
	}
	return false
}

// IsQuotaEnabled returns true if bitrot is enabled for a volume and false otherwise
func IsQuotaEnabled(v *Volinfo) bool {
	val, exists := v.Options[VkeyFeaturesQuota]
	if exists && val == "on" {
		return true
	}
	return false
}

//CheckBricksStatus will give detailed information about brick
func CheckBricksStatus(volinfo *Volinfo) ([]brick.Brickstatus, error) {

	var brickStatuses []brick.Brickstatus
	mtabEntries, err := getMounts()
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
				s.Port = pmap.RegistrySearch(binfo.Path, pmap.GfPmapPortBrickserver)
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
			if strings.HasPrefix(binfo.Path, m.mntDir) {
				s.MountOpts = m.mntOpts
				s.Device = m.fsName
				s.FS = m.mntType
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
	return "", errors.New("Failed To Get Mount Root")
}

//GetBrickMountDevice return device name of the mount point
func GetBrickMountDevice(brickPath, mountRoot string) (string, error) {
	mtabEntries, err := getMounts()
	if err != nil {
		return "", err
	}

	for _, entry := range mtabEntries {
		if entry.mntDir == mountRoot {
			return entry.fsName, nil
		}
	}
	return "", errors.New("Mount Point Not Found")

}

//CreateSubvolInfo parses subvol  information for response
func CreateSubvolInfo(sv *[]Subvol) []api.Subvol {
	var subvols []api.Subvol

	for _, subvol := range *sv {
		var blist []api.BrickInfo
		for _, b := range subvol.Bricks {
			blist = append(blist, brick.CreateBrickInfo(&b))
		}

		subvols = append(subvols, api.Subvol{
			Name:         subvol.Name,
			Type:         api.SubvolType(subvol.Type),
			Bricks:       blist,
			ReplicaCount: subvol.ReplicaCount,
			ArbiterCount: subvol.ArbiterCount,
		})
	}
	return subvols
}

//CreateVolumeInfoResp parses volume  information for response
func CreateVolumeInfoResp(v *Volinfo) *api.VolumeInfo {

	return &api.VolumeInfo{
		ID:        v.ID,
		Name:      v.Name,
		Type:      api.VolType(v.Type),
		Transport: v.Transport,
		DistCount: v.DistCount,
		State:     api.VolState(v.State),
		Options:   v.Options,
		Subvols:   CreateSubvolInfo(&v.Subvols),
	}
}
