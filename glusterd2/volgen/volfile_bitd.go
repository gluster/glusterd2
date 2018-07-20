package volgen

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateBitdVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, peerid uuid.UUID) {
	volfile.FileName = "gluster/bitd"

	bitd := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("bitd")

	for volIdx, vol := range clusterinfo {
		if vol.State != volume.VolStarted {
			continue
		}

		//If bitrot not enabled for volume, then skip those bricks
		val, exists := vol.Options["bitrot-stub.bitrot"]
		if exists && val == "on" {
			name := fmt.Sprintf("%s-bit-rot-%d", vol.Name, volIdx)
			bitdvol := bitd.Add("features/bit-rot", vol, nil).SetName(name).SetIgnoreOptions([]string{"scrubber"})
			clusterGraph(volfile, bitdvol, vol, peerid, &clusterGraphFilters{onlyLocalBricks: true})
		}
	}
}

func init() {
	registerClusterVolfile("bitd", generateBitdVolfile, true)
}
