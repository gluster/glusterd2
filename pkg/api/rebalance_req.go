package api

import "github.com/pborman/uuid"

// RebalanceStatus represents Rebalance Status
type RebalanceStatus uint64

const (
	// Started should be set only for a volume that has been just started rebalance process
	Started RebalanceStatus = iota
	// Inprogress should be set only for a volume that are running rebalance process
	Inprogress
	// Failed should be set only for a volume that are failed to run rebalance process
	Failed
	// Completed sholud be set only for a volume that the rebalance process is completed
	Completed
)

// RebalanceInfo represents a Rebalance Info Details
type RebalanceInfo struct {
	Volname     string          `json:"volname"`
	Status      RebalanceStatus `json:"status"`
	RebalanceID uuid.UUID       `json:"rebalanceID"`
}

// RebalanceStartResp represents Rebalance Start Response
type RebalanceStartResp RebalanceInfo
