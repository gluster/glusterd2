package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/volume"
)

var shdKeys = [...]string{"afr.self-heal-daemon", "replicate.self-heal-daemon"}

// isVolReplicate returns true if volume is of type replicate, disperse, distreplicate or distdisperse
// otherwise it returns false
func isVolReplicate(vType volume.VolType) bool {
	if vType == volume.Replicate || vType == volume.Disperse || vType == volume.DistReplicate || vType == volume.DistDisperse {
		return true
	}

	return false
}

// isHealEnabled returns true if heal is enabled for the volume otherwise returns false.
func isHealEnabled(v *volume.Volinfo) bool {
	for _, key := range shdKeys {
		value, ok := v.Options[key]
		if ok && value == "on" {
			return true
		}
	}
	return false
}
