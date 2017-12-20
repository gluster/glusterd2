package api

import "github.com/pborman/uuid"

// RebalanceStatus represents Rebalance Status
type RebalanceStatus uint64

// RebalanceInfo represents a Rebalance Info Details to create rebalance info
type RebalanceInfo struct {
	Volname     string          `json:"volname"`
	Status      RebalanceStatus `json:"status"`
	RebalanceID uuid.UUID       `json:"rebalanceID"`
}

// RebalanceStartResp represents Rebalance Start Response
type RebalanceStartResp RebalanceInfo
