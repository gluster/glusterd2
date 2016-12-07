package utils

import (
	"errors"
	"net"
	"strconv"
)

const (
	// DefaultRPCPort is the port on which glusterd2 instances will
	// listen for incoming RPC requests when user hasn't explicitly
	// set the port via config option.
	DefaultRPCPort = 24008
	// Accessing this as utils.DefaultRPCPort isn't neat. Just doing so
	// to avoid cyclic imports for now. We may need a separate package
	// other than 'gdctx' for global constants.
)

// FormRemotePeerAddress will check and validate peeraddress provided. It will
// return an address of the form <ip:port>
func FormRemotePeerAddress(peeraddress string) (string, error) {
	var remotePeerAddress string

	host, port, err := net.SplitHostPort(peeraddress)
	if err != nil || host == "" {
		return "", errors.New("Invalid peer address")
	}

	if port == "" {
		// Assume remote glusterd2 instance is using default rpc port
		remotePeerAddress = host + ":" + strconv.Itoa(DefaultRPCPort)
	} else {
		// If the address contains a port, just use that
		remotePeerAddress = peeraddress
	}

	return remotePeerAddress, nil
}
