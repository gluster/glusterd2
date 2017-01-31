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

	log "github.com/Sirupsen/logrus"
)

var programsList []Program

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

	port := getPortFromListener(listener)

	programsList = append(programsList, newGfHandshake())
	programsList = append(programsList, newGfDump())
	programsList = append(programsList, newGfPortmap())

	// Register all programs
	for _, program := range programsList {
		err := registerProgram(server, program, port)
		if err != nil {
			log.WithError(err).Error("Could not register RPC program " + program.Name())
			return err
		}
	}

	go func() {
		for {
			// TODO: The net.Conn instance here should be exposed
			// externally for:
			//     1. Sending notifications to glusterfs clients
			//        (example: volfile changed)
			//     2. Tracking number of glusterfs clients that
			//        are connected to glusterd2.
			// Multiple goroutines can safely invoke write on an
			// instance of net.Conn simultaneously.
			conn, err := listener.Accept()
			if err != nil {
				// TODO: Handle error ?
				return
			}
			log.WithField("address", conn.RemoteAddr().String()).Info("glusterfs client connected")
			go server.ServeCodec(sunrpc.NewServerCodec(conn))
		}
	}()

	return nil
}
