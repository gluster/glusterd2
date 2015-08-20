package errors

import (
	"errors"
)

// Different error macros
var (
	ErrVolNotFound       = errors.New("volume not found")
	ErrJSONParsingFailed = errors.New("unable to parse the request")
	ErrEmptyVolName      = errors.New("volume name is empty")
	ErrEmptyBrickList    = errors.New("brick list is empty")
	ErrVolExists         = errors.New("volume already exists")
	ErrVolAlreadyStarted = errors.New("volume already started")
	ErrVolAlreadyStopped = errors.New("volume already stopped")
)
