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
