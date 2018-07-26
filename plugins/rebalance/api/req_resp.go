package api

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
	// CmdNone indicates an invalid command
	CmdNone Command = iota
	// CmdStart starts the rebalance
	CmdStart
	// CmdStop stops the rebalance
	CmdStop
	// CmdStatus gets the rebalance status
	CmdStatus
	// CmdFixLayoutStart starts a rebalance fix-layout operation
	CmdFixLayoutStart
	// CmdStartForce starts rebalance with the force option
	CmdStartForce
)

// RebalNodeStatus represents the rebalance status on the Node
type RebalNodeStatus struct {
	PeerID            uuid.UUID `json:"peerid"`
	Status            string    `json:"status"`
	RebalancedFiles   string    `json:"rebalanced-files"`
	RebalancedSize    string    `json:"size"`
	LookedupFiles     string    `json:"lookedup"`
	SkippedFiles      string    `json:"skipped"`
	RebalanceFailures string    `json:"failed"`
	ElapsedTime       string    `json:"run-time"`
	TimeLeft          string    `json:"time-left"`
}

// RebalInfo represents the rebalance operation information
type RebalInfo struct {
	Volname     string
	State       Status
	Cmd         Command
	RebalanceID uuid.UUID
	CommitHash  uint64
	RebalStats  []RebalNodeStatus
}

// RebalStatus represents the rebalance status response
type RebalStatus struct {
	Volname     string            `json:"volume"`
	RebalanceID uuid.UUID         `json:"rebalance-id"`
	Nodes       []RebalNodeStatus `json:"nodes-status"`
}

// StartReq contains the options passed to the Rebalance Start Request
type StartReq struct {
	Option string `json:"option,omitempty"`
}
