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
	// CmdStop should be set only when given cmd is rebalance stop
	CmdStop
	// CmdStatus should be set only when given cmd is rebalance status
	CmdStatus
	// CmdFixlayoutStart should be set only when given cmd is rebalance fix-layout start
	CmdFixLayoutStart
	// CmdStartForce should be set only when given cmd is rebalance start force
	CmdStartForce
)

// NodeInfo represents Node information needed to store
type RebalNodeStatus struct {
	NodeID            uuid.UUID
	Status            string // This is per Node Rebalance daemon Status
	RebalancedFiles   string
	RebalancedSize    string
	LookedupFiles     string
	SkippedFiles      string
	RebalanceFailures string
	ElapsedTime       string
	TimeLeft          string
}

// RebalInfo represents Rebalance details for a node
type RebalInfo struct {
	Volname     string
	Status      Status
	Cmd         Command
	RebalanceID uuid.UUID
	CommitHash  uint64
	RebalStats  []RebalNodeStatus
}

// RebalInfo represents Rebalance details for a node
type RebalStatus struct {
	Volname     string
	RebalanceID uuid.UUID
	Nodes       []RebalNodeStatus
}

// StartReq represents Rebalance Start Request
type StartReq struct {
	Option string `omitempty`
}
