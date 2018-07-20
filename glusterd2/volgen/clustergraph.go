package volgen

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

const (
	thinArbiterOptionName  = "replicate.thin-arbiter"
	thinArbiterDefaultPort = "24007"
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

func clusterGraph(volfile *Volfile, dht *Entry, vol *volume.Volinfo, peerid uuid.UUID, filters *clusterGraphFilters) {
	numSubvols := len(vol.Subvols)
	decommissionedBricks := []string{}
	clientIdx := 0

	for _, subvol := range vol.Subvols {
		var parent *Entry
		var afrPendingXattr []string

		if filters != nil {
			if filters.noSubvolParent {
				// If No separate parent required for all the bricks
				parent = dht
			} else if len(filters.subvolTypes) > 0 && !typeInSubvolType(subvol.Type, filters.subvolTypes) {
				// If Graph need to be generated only for specific Subvolume Types
				continue
			}
		}

		// If Graph is required only for Local Bricks then
		// do not include sub volume in graph if no local bricks
		// exists for the sub volume
		if filters != nil && filters.onlyLocalBricks && len(subvol.GetLocalBricks()) == 0 {
			continue
		}

		// If Not set in prev filter checks
		if parent == nil {
			switch subvol.Type {
			case volume.SubvolReplicate:
				parent = dht.Add("cluster/replicate", vol, nil).SetName(subvol.Name)
			case volume.SubvolDisperse:
				parent = dht.Add("cluster/disperse", vol, nil).SetName(subvol.Name)
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
				afrPendingXattr = append(afrPendingXattr, fmt.Sprintf("%s-client-%d", vol.Name, clientIdx))
				clientIdx++

				// If local bricks only
				if filters != nil && filters.onlyLocalBricks && !uuid.Equal(b.PeerID, peerid) {
					continue
				}

				name := fmt.Sprintf("%s-client-%d", subvol.Name, brickIdx)
				parent.Add("protocol/client", vol, &b).SetName(name)
				if b.Decommissioned {
					decommissionedBricks = append(decommissionedBricks, name)
				}
			}

			thinarbiter, exists := vol.Options[thinArbiterOptionName]
			if exists && thinarbiter != "" {
				taParts := strings.Split(thinarbiter, ":")
				if len(taParts) != 2 && len(taParts) != 3 {
					// Thin Arbiter Option may not be valid
					continue
				}

				remotePort := thinArbiterDefaultPort
				if len(taParts) >= 3 {
					remotePort = taParts[2]
				}

				afrPendingXattr = append(afrPendingXattr, fmt.Sprintf("%s-ta-%d", vol.Name, clientIdx))
				clientIdx++

				name := fmt.Sprintf("%s-thin-arbiter-client", subvol.Name)
				parent.Add("protocol/client", vol, nil).
					SetName(name).
					SetExtraData(map[string]string{
						"brick.hostname": taParts[0],
						"brick.path":     taParts[1],
					}).
					SetExtraOptions(map[string]string{
						"remote-port": remotePort,
					})
			}

			if subvol.Type == volume.SubvolReplicate || subvol.Type == volume.SubvolDisperse {
				extraopts := make(map[string]string)
				if volfile.Name == "glustershd" {
					extraopts["iam-self-heal-daemon"] = "yes"
				}

				// Below option is required for shd and client volfile
				if volfile.Name == "glustershd" || volfile.Name == "fuse" {
					extraopts["afr-pending-xattr"] = strings.Join(afrPendingXattr, ",")
				}

				parent.SetExtraOptions(extraopts)
			}
		}
	}

	if len(decommissionedBricks) > 0 && volfile.Name == "rebalance" {
		dht.SetExtraOptions(map[string]string{
			"decommissioned-bricks": strings.Join(decommissionedBricks, " "),
		})
	}
}
