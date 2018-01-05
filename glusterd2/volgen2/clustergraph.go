package volgen2

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

type clusterGraphFilters struct {
	onlyLocalBricks bool
	noSubvolParent  bool
	subvolTypes     []volume.SubvolType
}

func typeInSubvolType(ele volume.SubvolType, list []volume.SubvolType) bool {
	for _, b := range list {
		if b == ele {
			return true
		}
	}
	return false
}

func clusterGraph(dht *Entry, vol *volume.Volinfo, nodeid uuid.UUID, filters *clusterGraphFilters) {
	numSubvols := len(vol.Subvols)
	for _, subvol := range vol.Subvols {
		var parent *Entry

		if filters != nil {
			if filters.noSubvolParent {
				// If No separate parent required for all the bricks
				parent = dht
			} else if len(filters.subvolTypes) > 0 && !typeInSubvolType(subvol.Type, filters.subvolTypes) {
				// If Graph need to be generated only for specific Subvolume Types
				continue
			}
		}

		// If Not set in prev filter checks
		if parent == nil {
			switch subvol.Type {
			case volume.SubvolReplicate:
				parent = dht.Add("cluster/replicate", vol, nil).SetName(subvol.Name)
			case volume.SubvolDistribute:
				if numSubvols == 1 {
					parent = dht
				} else {
					parent = dht.Add("cluster/distribute", vol, nil).SetName(subvol.Name)
				}
			default:
				parent = nil
			}
		}

		if parent != nil {
			for brickIdx, b := range subvol.Bricks {
				// If local bricks only
				if filters != nil && filters.onlyLocalBricks && !uuid.Equal(b.NodeID, nodeid) {
					continue
				}

				name := fmt.Sprintf("%s-client-%d", subvol.Name, brickIdx)
				parent.Add("protocol/client", vol, &b).SetName(name)
			}
		}
	}
}
