package errors

import (
	"errors"
)

// Different error macros
var (
	ErrVolCreateFail           = errors.New("unable to create volume")
	ErrVolNotFound             = errors.New("volume not found")
	ErrJSONParsingFailed       = errors.New("unable to parse the request")
	ErrEmptyVolName            = errors.New("volume name is empty")
	ErrEmptyBrickList          = errors.New("brick list is empty")
	ErrInvalidBrickPath        = errors.New("invalid brick path")
	ErrVolExists               = errors.New("volume already exists")
	ErrVolAlreadyStarted       = errors.New("volume already started")
	ErrVolAlreadyStopped       = errors.New("volume already stopped")
	ErrWrongGraphType          = errors.New("graph: incorrect graph type")
	ErrDeviceIDNotFound        = errors.New("Failed to get device id")
	ErrBrickIsMountPoint       = errors.New("Brick path is already a mount point")
	ErrBrickUnderRootPartition = errors.New("Brick path is under root partition")
	ErrBrickNotDirectory       = errors.New("Brick path is not a directory")
)
