package afr

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

var names = [...]string{"replicate", "afr"}

// isVolReplicate returns true if volume is of type replicate, disperse, distreplicate or distdisperse
// otherwise it returns false
func isVolReplicate(vType volume.VolType) bool {
	if vType == volume.Replicate || vType == volume.Disperse || vType == volume.DistReplicate || vType == volume.DistDisperse {
		return true
	}

	return false
}

func validateOptions(v *volume.Volinfo, key string, value string) error {
	switch key {
	case "metadata-self-heal":
		if v.Subvols[0].ReplicaCount == 1 {
			return errors.New("option cannot be set for a non replicate volume")
		}
	case "self-heal-daemon":
		if !isVolReplicate(v.Type) {
			return errors.New("option cannot be set for a non replicate volume")
		}
	}
	return nil
}

func init() {
	for _, name := range names {
		xlator.RegisterValidationFunc(name, validateOptions)
	}
}
