package utils

import (
	"errors"
	"net"
	"strings"

	config "github.com/spf13/viper"
)

// FormRemotePeerAddress will check and validate peeraddress provided. It will
// return an address of the form <host:port>
func FormRemotePeerAddress(peeraddress string) (string, error) {

	host, port, err := net.SplitHostPort(peeraddress)
	if err != nil {
		// net.SplitHostPort() returns an error if port is missing.
		if strings.HasSuffix(err.Error(), "missing port in address") {
			host = peeraddress
			port = config.GetString("defaultpeerport")
		} else {
			return "", err
		}
	}

	if host == "" {
		return "", errors.New("invalid peer address")
	}

	remotePeerAddress := host + ":" + port
	return remotePeerAddress, nil
}

// IsPeerAddressSame checks if two peer addresses are same by normalizing
// each address to <ip>:<port> form.
func IsPeerAddressSame(addr1 string, addr2 string) bool {
	r1, _ := FormRemotePeerAddress(addr1)
	r2, _ := FormRemotePeerAddress(addr2)
	return r1 == r2
}
