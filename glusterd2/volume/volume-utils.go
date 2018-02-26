package volume

import (
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/pkg/errors"

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
				return errors.ErrBrickPathAlreadyInUse
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
			s.Size = *(createSizeInfo(&fstat))
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
