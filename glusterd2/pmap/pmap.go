package pmap

import (
	"net"
)

// RegistrySearch searches for a brick path in the pmap registry and
// returns the port assigned to it.
func RegistrySearch(brickpath string) (int, error) {
	return registry.SearchByBrickPath(brickpath)
}

// ProcessDisconnect will handle a TCP connection disconnection
func ProcessDisconnect(conn net.Conn) error {
	return registry.RemovePortByConn(conn)
}
