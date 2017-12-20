package volgen2

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateShdVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gluster/glustershd"
	shd := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("glustershd")

	for _, vol := range clusterinfo {
		for subvolIdx, subvol := range vol.Subvols {
			if subvol.Type == volume.SubvolReplicate {
				name := fmt.Sprintf("%s-replicate-%d", vol.Name, subvolIdx)
				replicate := shd.Add("cluster/replicate", vol, nil).SetName(name)
				for brickIdx, b := range subvol.Bricks {
					name := fmt.Sprintf("%s-replicate-%d-client-%d", vol.Name, subvolIdx, brickIdx)
					replicate.Add("protocol/client", vol, &b).SetName(name)
				}
			}
		}
	}
}

func init() {
	registerClusterVolfile("glustershd", generateShdVolfile, false)
}
