package volumecommands

import (
	"errors"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
)

func nodesFromVolumeCreateReq(req *api.VolCreateReq) ([]uuid.UUID, error) {
	var nodesMap = make(map[string]int)
	var nodes []uuid.UUID
	for _, subvol := range req.Subvols {
		for _, brick := range subvol.Bricks {
			if _, ok := nodesMap[brick.NodeID]; !ok {
				nodesMap[brick.NodeID] = 1
				u := uuid.Parse(brick.NodeID)
				if u == nil {
					return nil, errors.New("Unable to parse Node ID")
				}
				nodes = append(nodes, u)
			}
		}
	}
	return nodes, nil
}

func nodesFromVolumeExpandReq(req *api.VolExpandReq) ([]uuid.UUID, error) {
	var nodesMap = make(map[string]int)
	var nodes []uuid.UUID
	for _, brick := range req.Bricks {
		if _, ok := nodesMap[brick.NodeID]; !ok {
			nodesMap[brick.NodeID] = 1
			u := uuid.Parse(brick.NodeID)
			if u == nil {
				return nil, errors.New("Unable to parse Node ID")
			}
			nodes = append(nodes, u)
		}
	}
	return nodes, nil
}
