package volume

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

var (
	volumeNameRE = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
)

// GenerateVolumeName generates volume name as vol_<random-id>
func GenerateVolumeName() string {
	return "vol_" + uuid.NewRandom().String()
}

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
	volumes, e := GetVolumes()
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
			Name:          subvol.Name,
			Type:          api.SubvolType(subvol.Type),
			Bricks:        blist,
			ReplicaCount:  subvol.ReplicaCount,
			ArbiterCount:  subvol.ArbiterCount,
			DisperseCount: subvol.DisperseCount,
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

	return resp
}
