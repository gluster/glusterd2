package rebalance

import (
	"errors"
)

var (
	// ErrVolNotDistribute : Cannot run rebalance on a non distribute volume
	ErrVolNotDistribute = errors.New("Not a distribute volume")
	// ErrRebalanceNotStarted : Rebalance not started on the volume
	ErrRebalanceNotStarted    = errors.New("Rebalance not started")
	ErrRebalanceInvalidOption = errors.New("Invalid Rebalance start option")
)
