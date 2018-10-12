package deviceutils

import (
	"errors"
)

var (
	// ErrDeviceNotFound represents device not found error
	ErrDeviceNotFound = errors.New("device does not exist in the given peer")
)
