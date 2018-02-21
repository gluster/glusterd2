package volgen2

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateScrubVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gluster/scrub"

	scrub := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("scrub")

	for volIdx, vol := range clusterinfo {
		//If bitrot not enabled for volume, then skip those bricks
		val, exists := vol.Options[volume.VkeyFeaturesBitrot]
		if exists && val == "on" {
			name := fmt.Sprintf("%s-bit-rot-%d", vol.Name, volIdx)
			scrubvol := scrub.Add("features/bit-rot", vol, nil).SetName(name)
			clusterGraph(volfile, scrubvol, vol, nodeid, &clusterGraphFilters{onlyLocalBricks: true})
		}
	}
}

func init() {
	registerClusterVolfile("scrub", generateScrubVolfile, true)
}
