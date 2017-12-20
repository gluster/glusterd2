package volgen2

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateBitdVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gluster/bitd"

	bitd := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("bitd")

	for volIdx, vol := range clusterinfo {
		// TODO: If bitrot not enabled for volume, then skip
		name := fmt.Sprintf("%s-bit-rot-%d", vol.Name, volIdx)
		bitdvol := bitd.Add("features/bit-rot", vol, nil).SetName(name)
		for subvolIdx, subvol := range vol.Subvols {
			for brickIdx, b := range subvol.Bricks {
				// Only include local bricks
				if !uuid.Equal(b.NodeID, nodeid) {
					continue
				}
				name := fmt.Sprintf("%s-%d-client-%d", vol.Name, subvolIdx, brickIdx)
				bitdvol.Add("protocol/client", vol, &b).SetName(name)
			}
		}
	}
}

func init() {
	registerClusterVolfile("bitd", generateBitdVolfile, true)
}
