package sunrpc

import (
	"expvar"
	"io"
	"net"
	"net/rpc"
	"strconv"
	"sync"

	"github.com/gluster/glusterd2/bin/glusterd2/pmap"
	"github.com/gluster/glusterd2/plugins"

	"github.com/prashanthpai/sunrpc"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

var (
	// metrics
	clientCount = expvar.NewInt("sunrpc_clients_connected")
)

// SunRPC implements a suture service
type SunRPC struct {
	server   *rpc.Server
	listener net.Listener
	stopCh   chan struct{}
}

var programsList []sunrpc.Program

var clientsList = struct {
	sync.RWMutex
	c map[net.Conn]bool
}{
	// This map is used as a set. Values are not consumed.
	c: make(map[net.Conn]bool),
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

// NewMuxed returns a SunRPC server configured to listen on a CMux multiplexed connection
func NewMuxed(m cmux.CMux) *SunRPC {

	srv := &SunRPC{
		server:   rpc.NewServer(),
		listener: m.Match(sunrpc.CmuxMatcher()),
		stopCh:   make(chan struct{}),
	}

	programsList = []sunrpc.Program{
		newGfHandshake(),
		newGfDump(),
		pmap.NewGfPortmap(),
	}

	for _, p := range plugins.PluginsList {
		rpcProcs := p.SunRPCProgram()
		if rpcProcs != nil {
			programsList = append(programsList, rpcProcs)
			log.WithField("plugin", p.Name()).Debug("loaded sunrpc procedures from plugin")
		}
	}

	port := getPortFromListener(srv.listener)

	for _, prog := range programsList {
		err := registerProgram(srv.server, prog, port, false)
		if err != nil {
			log.WithError(err).WithField("program", prog.Name()).Error("could not register SunRPC program")
			return nil
		}
	}

	return srv
}

// Serve will start accepting Sun RPC client connections on the listener
// provided.
func (s *SunRPC) Serve() {

	// Detect client disconnections
	notifyClose := make(chan io.ReadWriteCloser, 10)
	go func() {
		for rwc := range notifyClose {
			conn := rwc.(net.Conn)
			log.WithField("address", conn.RemoteAddr().String()).Info("sunrpc client disconnected")

			// Update list of clients
			clientsList.Lock()
			delete(clientsList.c, conn)
			clientsList.Unlock()
			clientCount.Add(-1)
		}
	}()

	log.WithField("ip:port", s.listener.Addr().String()).Info("started GlusterD SunRPC server")
	for {
		select {
		case <-s.stopCh:
			// TODO: Gracefully stop the server: https://github.com/golang/go/issues/17239
			log.Debug("stopping GlusterD SunRPC server")
			// We have 3 nested cmux listeners - cmux, rest and sunrpc.
			// Closing listener in cmux and rest server.
			log.Info("stopped GlusterD SunRPC server")
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if err != cmux.ErrListenerClosed {
				log.WithError(err).Error("failed to accept incoming connection")
			}
			continue
		}

		clientCount.Add(1)

		// Update list of clients
		clientsList.Lock()
		clientsList.c[conn] = true
		clientsList.Unlock()
		log.WithField("address", conn.RemoteAddr().String()).Info("sunrpc client connected")

		go s.server.ServeCodec(sunrpc.NewServerCodec(conn, notifyClose))
	}
}

// Stop stops the SunRPC server
func (s *SunRPC) Stop() {
	close(s.stopCh)
}
