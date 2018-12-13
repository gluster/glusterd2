package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/volume"
)

const (
	selfHealKey          = "self-heal-daemon"
	shdKey               = "cluster/replicate." + selfHealKey
	granularEntryHealKey = "granular-entry-heal"
)

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
	value, ok := v.Options[shdKey]
	if ok && value == "on" {
		return true
	}
	return false
}
