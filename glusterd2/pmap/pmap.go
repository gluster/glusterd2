package pmap

import "net"

// AssignPort assigns an available port to brick specified by the brick path
// and returns the port number allocated. Returns error if it cannot allocate
// a free port.
func AssignPort(brickpath string) (int, error) {
	return registry.Allocate(brickpath)
}

// RegistrySearch searches for a brick path in the pmap registry and
// returns the port assigned to it.
func RegistrySearch(brickpath string) (int, error) {
	return registry.SearchByBrickPath(brickpath)
}

// ProcessDisconnect will handle a TCP connection disconnection
func ProcessDisconnect(conn net.Conn) error {
	return registry.RemovePortByConn(conn)
}
