package utils

import (
	"errors"
	"net"

	config "github.com/spf13/viper"
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
		remotePeerAddress = host + ":" + config.GetString("defaultrpcport")
	} else {
		// If the address contains a port, just use that
		remotePeerAddress = peeraddress
	}

	return remotePeerAddress, nil
}
