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

func IsPortFree(port int) bool {
	// TODO: Ports can be bound to specific interfaces. This function
	// should be modified to take a IP/hostname arg as well.

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	} else {
		conn.Close()
		return true
	}
}

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
