package sunrpc

import (
	"net"
	"net/rpc"
	"strconv"

	"github.com/prashanthpai/sunrpc"

	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

type SunRPC struct {
	server   *rpc.Server
	listener net.Listener
	stop     chan bool
}

// New returns a SunRPC server configured to listen on the given listener
func New(l net.Listener) *SunRPC {
	srv := &SunRPC{
		server:   rpc.NewServer(),
		listener: l,
		stop:     make(chan bool, 1),
	}

	err := registerHandshakeProgram(srv.server, getPortFromListener(srv.listener))
	if err != nil {
		log.WithError(err).Error("Could not register handshake program")
		return nil
	}
	return srv
}

// NewMuxed returns a SunRPC server configured to listen on a CMux multiplexed connection
func NewMuxed(m cmux.CMux) *SunRPC {
	return New(m.Match(sunrpc.CmuxMatcher()))
}

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

// Serve will start accepting Sun RPC client connections on the listener
// provided.
func (s *SunRPC) Serve() {
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		// TODO: The net.Conn instance here should be exposed
		// externally for:
		//     1. Sending notifications to glusterfs clients
		//        (example: volfile changed)
		//     2. Tracking number of glusterfs clients that
		//        are connected to glusterd2.
		// Multiple goroutines can safely invoke write on an
		// instance of net.Conn simultaneously.
		conn, err := s.listener.Accept()
		if err != nil {
			// TODO: Handle error ?
			continue
		}
		log.WithField("address", conn.RemoteAddr().String()).Info("glusterfs client connected")
		go s.server.ServeRequest(sunrpc.NewServerCodec(conn))
	}
	return
}

// Stop stops the SunRPC server
func (s *SunRPC) Stop() {
	close(s.stop)
	// TODO: Gracefully stop the server
}
