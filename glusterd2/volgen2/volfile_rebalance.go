package volgen2

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateRebalanceVolfile(volfile *Volfile, vol *volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "rebalance/" + vol.Name

	dht := volfile.RootEntry.Add("debug/io-stats", vol, nil).SetName(vol.Name).
		Add("cluster/distribute", vol, nil)

	for subvolIdx, subvol := range vol.Subvols {
		if subvol.Type == volume.SubvolReplicate {
			name := fmt.Sprintf("%s-replicate-%d", vol.Name, subvolIdx)
			replicate := dht.Add("cluster/replicate", vol, nil).SetName(name)

			for brickIdx, b := range subvol.Bricks {
				name := fmt.Sprintf("%s-replicate-%d-client-%d", vol.Name, subvolIdx, brickIdx)
				replicate.Add("protocol/client", vol, &b).SetName(name)
			}
		}
	}
}

func init() {
	registerVolumeVolfile("rebalance", generateRebalanceVolfile, false)
}
