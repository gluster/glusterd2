package pmap

import (
	"fmt"
	"net"
)

func isPortFree(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
