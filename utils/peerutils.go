package utils

import (
	"errors"
	"net"
	"strings"

	config "github.com/spf13/viper"
)

// FormRemotePeerAddress will check and validate peeraddress provided. It will
// return an address of the form <ip:port>
func FormRemotePeerAddress(peeraddress string) (string, error) {

	host, port, err := net.SplitHostPort(peeraddress)
	if err != nil {
		// net.SplitHostPort() returns an error if port is missing.
		if strings.HasPrefix(err.Error(), "missing port in address") {
			host = peeraddress
			port = config.GetString("defaultrpcport")
		} else {
			return "", err
		}
	}

	if host == "" {
		return "", errors.New("Invalid peer address")
	}

	remotePeerAddress := host + ":" + port
	return remotePeerAddress, nil
}
