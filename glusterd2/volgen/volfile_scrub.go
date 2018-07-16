package volgen

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateScrubVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, peerid uuid.UUID) {
	volfile.FileName = "gluster/scrub"

	scrub := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("scrub")

	for volIdx, vol := range clusterinfo {
		if vol.State != volume.VolStarted {
			continue
		}

		//If bitrot not enabled for volume, then skip those bricks
		val, exists := vol.Options["bitrot-stub.bitrot"]
		if exists && val == "on" {
			name := fmt.Sprintf("%s-bit-rot-%d", vol.Name, volIdx)
			scrubvol := scrub.Add("features/bit-rot", vol, nil).SetName(name)
			clusterGraph(volfile, scrubvol, vol, peerid, &clusterGraphFilters{onlyLocalBricks: true})
		}
	}
}

func init() {
	registerClusterVolfile("scrub", generateScrubVolfile, true)
}
