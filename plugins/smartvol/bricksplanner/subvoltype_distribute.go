package bricksplanner

import (
	smartvolapi "github.com/gluster/glusterd2/plugins/smartvol/api"
)

type distributeSubvolPlanner struct {
	subvolSize uint64
}

func (s *distributeSubvolPlanner) Init(req *smartvolapi.Volume, subvolSize uint64) {
	s.subvolSize = subvolSize
}

func (s *distributeSubvolPlanner) BricksCount() int {
	return 1
}

func (s *distributeSubvolPlanner) BrickSize(idx int) uint64 {
	return s.subvolSize
}

func (s *distributeSubvolPlanner) BrickType(idx int) string {
	return "Brick"
}

func init() {
	subvolPlanners["distribute"] = &distributeSubvolPlanner{}
}
