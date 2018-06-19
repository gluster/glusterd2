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
	ErrProgUnavail = errors.New("remote server has not exported program")
	ErrProcUnavail = errors.New("remote server has no such procedure")
	ErrGarbageArgs = errors.New("remote procedure cannot decode params")
	ErrSystemErr   = errors.New("system error on remote server")
)

// These errors represent invalid replies from server and auth rejection.
var (
	ErrInvalidRPCMessageType = errors.New("invalid RPC message type received")
	ErrInvalidRPCRepyType    = errors.New("invalid RPC reply received. Reply type should be MsgAccepted or MsgDenied")
	ErrInvalidMsgDeniedType  = errors.New("invalid MsgDenied reply. Possible values are RPCMismatch and AuthError")
	ErrInvalidMsgAccepted    = errors.New("invalid MsgAccepted reply received")
	ErrAuthError             = errors.New("remote server rejected identity of the caller")
)
