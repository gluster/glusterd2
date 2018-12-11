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

// RegistryExtend adds a brick entry to pmap registry and is used during
// multiplexing a brick.
func RegistryExtend(brickpath string, port int, pid int) {
	registry.Update(port, brickpath, nil, pid)
}

// GetBricksOnPort returns a list of bricks that are multiplexed onto a single
// process that is listening on the port specified.
func GetBricksOnPort(port int) []string {
	var bricks []string

	if m, ok := registry.Ports[port]; ok {
		for path := range m {
			bricks = append(bricks, path)
		}
	}

	return bricks
}
