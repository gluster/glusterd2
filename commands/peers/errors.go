package peercommands

const (
	// Errors returned during peer operations
	ErrNone PeerError = iota
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

type PeerError int32

func (p PeerError) String() string {
	if p <= ErrNone || p >= ErrMax {
		return "unknown error"
	}
	return errorStrings[p]
}

func (p PeerError) Error() string {
	return p.String()
}
