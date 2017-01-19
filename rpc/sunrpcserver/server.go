package sunrpcserver

/* TODO:
The rpc/* directory needs a more elegant subpackage structuring.
Something like shown below but without package names conflicting with
imported package names.

    server -
           |- grpc
           |- sunrpc
           |- rest

Each subpackage will implement the server for that protocol.
*/

import (
	"net"
	"net/rpc"
	"strconv"

	"github.com/prashanthpai/sunrpc"
)

func getPortFromListener(listener net.Listener) int {

	if listener == nil {
		return 0
	}

	addr := listener.Addr().String()
	_, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return 0
	}

	return port
}

// Start will start accepting Sun RPC client connections on the listener
// provided.
func Start(listener net.Listener) error {

	// There is no graceful shutdown of Sun RPC server yet. So this
	// instance doesn't have to be global variable or attached to gdctx.
	server := rpc.NewServer()

	err := registerHandshakeProgram(server, getPortFromListener(listener))
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeRequest(sunrpc.NewServerCodec(conn))
		}
	}()

	return nil
}
