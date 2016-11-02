package brick

// Consider moving this to utils package if found useful elsewhere

import (
	"fmt"
	"net"
)

const (
	minPort = 49152
	maxPort = 65535
)

// IsPortFree returns if the specified port is free or not.
func IsPortFree(port int) bool {
	// TODO: Ports can be bound to specific interfaces. This function
	// should be modified to take a IP/hostname arg as well.

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetNextAvailableFreePort returns the next available free port in the range minPort to maxPort
func GetNextAvailableFreePort() int {

	var p int

	for p = minPort; p < maxPort; p++ {
		if IsPortFree(p) == true {
			// Caller should handle TOCTOU races ?
			break
		}
	}

	return p
}
