package volume

import (
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
		for _, subvol := range v.Subvols {
			for _, b := range subvol.Bricks {
				if uuid.Equal(b.NodeID, nodeID) && b.Path == brickPath {
					log.Error("Brick is already used by ", v.Name)
					return errors.ErrBrickPathAlreadyInUse
				}
			}
		}
	}
	return nil
}
