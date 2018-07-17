package bricksplanner

import (
	"github.com/gluster/glusterd2/pkg/api"
)

type replicaSubvolPlanner struct {
	subvolSize       uint64
	replicaCount     int
	arbiterCount     int
	brickSize        uint64
	arbiterBrickSize uint64
}

func (s *replicaSubvolPlanner) Init(req *api.VolCreateReq, subvolSize uint64) {
	s.subvolSize = subvolSize
	s.replicaCount = req.ReplicaCount
	s.arbiterCount = req.ArbiterCount
	s.brickSize = s.subvolSize
	// TODO: Calculate Arbiter brick size
	s.arbiterBrickSize = s.subvolSize
}

func (s *replicaSubvolPlanner) BricksCount() int {
	return s.replicaCount + s.arbiterCount
}

func (s *replicaSubvolPlanner) BrickSize(idx int) uint64 {
	if idx == (s.replicaCount) && s.arbiterCount > 0 {
		return s.arbiterBrickSize
	}

	return s.brickSize
}

func (s *replicaSubvolPlanner) BrickType(idx int) string {
	if idx == (s.replicaCount) && s.arbiterCount > 0 {
		return "arbiter"
	}

	return "brick"
}

func init() {
	subvolPlanners["replicate"] = &replicaSubvolPlanner{}
}
