package rebalance

import (
	"errors"
)

var (
	// ErrVolNotDistribute : Cannot run rebalance on a non distribute volume
	ErrVolNotDistribute = errors.New("not a distribute volume")
	// ErrRebalanceNotStarted : Rebalance not started on the volume
	ErrRebalanceNotStarted = errors.New("rebalance not started")
	// ErrRebalanceInvalidOption : Invalid option provided to the rebalance start command
	ErrRebalanceInvalidOption = errors.New("invalid Rebalance start option")
)
