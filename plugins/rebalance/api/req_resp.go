package rebalance

import (
	"github.com/pborman/uuid"
)

// Status represents Rebalance Status
type Status uint64

const (
	// NotStarted should be set only for a node in which rebalance process is not started
	NotStarted Status = iota
	// Started should be set only for a node that has been just started rebalance process
	Started
	// Stopped should be set only for a node that has been just stopped rebalance process
	Stopped
	// Complete should be set only for a node that the rebalance process is completed
	Complete
	// Failed should be set only for a node that are failed to run rebalance process
	Failed
)

// Command represents Rebalance Commands
type Command uint64

const (
	// CmdNone should be set only when given cmd is none
	CmdNone Command = iota
	// CmdStart should be set only when given cmd is rebalance start
	CmdStart
	// CmdFixlayoutStart should be set only when given cmd is rebalance fix-layout start
	CmdFixlayoutStart
	// CmdStartForce should be set only when given cmd is rebalance start force
	CmdStartForce
	// CmdStop should be set only when given cmd is rebalance stop
	CmdStop
	// CmdStatus should be set only when given cmd is rebalance status
	CmdStatus
)

// NodeInfo represents Node information needed to store
type NodeInfo struct {
	NodeID            uuid.UUID
	Status            string // This is per Node Rebalance daemon Status
	RebalanceFiles    uint64
	RebalanceSize     uint64
	LookedupFiles     uint64
	RebalanceFailures uint64
	ElapsedTime       uint64
	SkippedFiles      uint64
	TimeLeft          uint64
}

// RebalInfo represents Rebalance details
type RebalInfo struct {
	Volname     string
	Status      Status // High level status, static information Started/Stopped etc
	Cmd         Command
	RebalanceID uuid.UUID
	CommitHash  uint64
	Nodes       []NodeInfo
}

// StartReq represents Rebalance Start Request
type StartReq struct {
	Fixlayout bool `omitempty`
	Force     bool `omitempty`
}
