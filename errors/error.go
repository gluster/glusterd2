package errors

import (
	"errors"
)

// Different error macros
var (
	ErrVolCreateFail           = errors.New("unable to create volume")
	ErrVolNotFound             = errors.New("volume not found")
	ErrPeerNotFound            = errors.New("peer not found")
	ErrJSONParsingFailed       = errors.New("unable to parse the request")
	ErrEmptyVolName            = errors.New("volume name is empty")
	ErrEmptyBrickList          = errors.New("brick list is empty")
	ErrInvalidBrickPath        = errors.New("invalid brick path, brick path should be in host:<brick> format")
	ErrVolExists               = errors.New("volume already exists")
	ErrVolAlreadyStarted       = errors.New("volume already started")
	ErrVolAlreadyStopped       = errors.New("volume already stopped")
	ErrWrongGraphType          = errors.New("graph: incorrect graph type")
	ErrDeviceIDNotFound        = errors.New("Failed to get device id")
	ErrBrickIsMountPoint       = errors.New("Brick path is already a mount point")
	ErrBrickUnderRootPartition = errors.New("Brick path is under root partition")
	ErrBrickNotDirectory       = errors.New("Brick path is not a directory")
	ErrBrickPathAlreadyInUse   = errors.New("Brick path is already in use by other gluster volume")
	ErrNoHostnamesPresent      = errors.New("no hostnames present")
	ErrBrickPathConvertFail    = errors.New("Failed to convert the brickpath to absolute path")
	ErrBrickNotLocal           = errors.New("Brickpath doesn't belong to localhost")
	ErrBrickPathTooLong        = errors.New("Brickpath too long")
	ErrSubDirPathTooLong       = errors.New("sub directory path is too long")
	ErrIPAddressNotFound       = errors.New("Failed to find IP address")
	ErrPeerLocalNode           = errors.New("The peer being added is the local node")
	ErrProcessNotFound         = errors.New("The process is not running or is inaccessible")
	ErrProcessAlreadyRunning   = errors.New("Process is already running")
)
