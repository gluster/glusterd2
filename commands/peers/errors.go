package peercommands

// Error is the error type returned by this package
type Error int32

// Errors returned by this package
// TODO: Add more errors
const (
	ErrNone Error = iota
	ErrAnotherCluster
	ErrHaveVolumes
	ErrStoreReconfigFailed
	ErrUnknownPeer
	ErrMax
)

var errorStrings [ErrMax]string

func init() {
	errorStrings[ErrNone] = "not an error"
	errorStrings[ErrAnotherCluster] = "peer is part of another cluster"
	errorStrings[ErrHaveVolumes] = "peer has existing volumes"
	errorStrings[ErrStoreReconfigFailed] = "store reconfigure failed on peer"
	errorStrings[ErrUnknownPeer] = "request recieved from unknown peer"
}

func (e Error) String() string {
	if e <= ErrNone || e >= ErrMax {
		return "unknown error"
	}
	return errorStrings[e]
}

func (e Error) Error() string {
	return e.String()
}
