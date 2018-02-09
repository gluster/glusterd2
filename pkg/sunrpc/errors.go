package sunrpc

import (
	"errors"
	"fmt"
)

// Internal errors
var (
	ErrInvalidFragmentSize    = errors.New("The RPC fragment size is invalid")
	ErrRPCMessageSizeExceeded = errors.New("The RPC message size is too big")
)

// RPC errors

// ErrRPCMismatch contains the lowest and highest version of RPC protocol
// supported by the remote server
type ErrRPCMismatch struct {
	Low  uint32
	High uint32
}

func (e ErrRPCMismatch) Error() string {
	return fmt.Sprintf("RPC version not supported by server. Lowest and highest supported versions are %d and %d respectively", e.Low, e.High)
}

// ErrProgMismatch contains the lowest and highest version of program version
// supported by the remote program
type ErrProgMismatch struct {
	Low  uint32
	High uint32
}

func (e ErrProgMismatch) Error() string {
	return fmt.Sprintf("Program version not supported. Lowest and highest supported versions are %d and %d respectively", e.Low, e.High)
}

// Given that the remote server accepted the RPC call, following errors
// represent error status of an attempt to call remote procedure
var (
	ErrProgUnavail = errors.New("Remote server has not exported program")
	ErrProcUnavail = errors.New("Remote server has no such procedure")
	ErrGarbageArgs = errors.New("Remote procedure cannot decode params")
	ErrSystemErr   = errors.New("System error on remote server")
)

// These errors represent invalid replies from server and auth rejection.
var (
	ErrInvalidRPCMessageType = errors.New("Invalid RPC message type received")
	ErrInvalidRPCRepyType    = errors.New("Invalid RPC reply received. Reply type should be MsgAccepted or MsgDenied")
	ErrInvalidMsgDeniedType  = errors.New("Invalid MsgDenied reply. Possible values are RPCMismatch and AuthError")
	ErrInvalidMsgAccepted    = errors.New("Invalid MsgAccepted reply received")
	ErrAuthError             = errors.New("Remote server rejected identity of the caller")
)
