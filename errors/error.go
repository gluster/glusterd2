package errors

import (
	"errors"
)

// Different error macros
var (
	ErrVolNotFound       = errors.New("Volume not found")
	ErrJSONParsingFailed = errors.New("Unable to parse the request")
	ErrEmptyVolName      = errors.New("Volume name is empty")
	ErrEmptyBrickList    = errors.New("Brick list is empty")
	ErrVolExists         = errors.New("Volume already exists")
	ErrVolAlreadyStarted = errors.New("Volume already started")
	ErrVolAlreadyStopped = errors.New("Volume already stopped")
)
