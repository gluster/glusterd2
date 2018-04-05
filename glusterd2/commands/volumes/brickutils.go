package volumecommands

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
)

func nodesFromVolumeCreateReq(req *api.VolCreateReq) ([]uuid.UUID, error) {
	var nodesMap = make(map[string]bool)
	var nodes []uuid.UUID
	for _, subvol := range req.Subvols {
		for _, brick := range subvol.Bricks {
			if _, ok := nodesMap[brick.PeerID]; !ok {
				nodesMap[brick.PeerID] = true
				u := uuid.Parse(brick.PeerID)
				if u == nil {
					return nil, fmt.Errorf("Failed to parse peer ID: %s", brick.PeerID)
				}
				nodes = append(nodes, u)
			}
		}
	}
	return nodes, nil
}

func nodesFromVolumeExpandReq(req *api.VolExpandReq) ([]uuid.UUID, error) {
	var nodesMap = make(map[string]bool)
	var nodes []uuid.UUID
	for _, brick := range req.Bricks {
		if _, ok := nodesMap[brick.PeerID]; !ok {
			nodesMap[brick.PeerID] = true
			u := uuid.Parse(brick.PeerID)
			if u == nil {
				return nil, fmt.Errorf("Failed to parse peer ID: %s", brick.PeerID)
			}
			nodes = append(nodes, u)
		}
	}
	return nodes, nil
}
