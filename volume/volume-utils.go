package volume

import (
	"os"
	"strings"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
)

var (
	getVolumesFunc = GetVolumes
)

// RemoveBrickPaths is to clean up the bricks in case commit fails for volume
// create
func RemoveBrickPaths(bricks []brick.Brickinfo) error {
	var e error
	for _, b := range bricks {
		local, err := utils.IsLocalAddress(b.Hostname)
		if err != nil || local == false {
			continue
		}
		err = os.Remove(b.Path)
		if err != nil {
			e := err
			log.WithFields(log.Fields{"error": e.Error(),
				"brickPath": b.Path,
				"host":      b.Hostname}).Error("Failed to remove directory")
		}
	}
	return e
}

// isBrickPathAvailable validates whether the brick is consumed by other
// volume
func isBrickPathAvailable(hostname string, brickPath string) error {
	volumes, e := getVolumesFunc()
	if e != nil || volumes == nil {
		// In case cluster doesn't have any volumes configured yet,
		// treat this as success
		log.Debug("Failed to retrieve volumes")
		return nil
	}
	for _, v := range volumes {
		for _, b := range v.Bricks {
			if b.Hostname == hostname && b.Path == brickPath {
				log.Error("Brick is already used by ", v.Name)
				return errors.ErrBrickPathAlreadyInUse
			}
		}
	}
	return nil
}

// SplitVolumeOptionName returns three strings by breaking volume option name
// of the form <graph>.<xlator>.<option> into its constituents. Specifying
// <graph> is optional and when omitted, the option change shall be applied to
// instances of the xlator loaded in all graphs.
func SplitVolumeOptionName(option string) (string, string, string) {
	tmp := strings.Split(strings.TrimSpace(option), ".")
	switch len(tmp) {
	case 2:
		return "", tmp[0], tmp[1]
	case 3:
		return tmp[0], tmp[1], tmp[2]
	}

	return "", "", ""
}
