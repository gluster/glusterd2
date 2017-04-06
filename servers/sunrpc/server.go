package sunrpc

import (
	"io"
	"net"
	"net/rpc"
	"reflect"
	"strconv"
	"sync"

	"github.com/gluster/glusterd2/plugins"
	"github.com/gluster/glusterd2/pmap"

	log "github.com/Sirupsen/logrus"
	"github.com/prashanthpai/sunrpc"
	"github.com/soheilhy/cmux"
)

// SunRPC implements a suture service
type SunRPC struct {
	server   *rpc.Server
	listener net.Listener
	stop     chan bool
}

var programsList []sunrpc.Program

var clientsList = struct {
	sync.RWMutex
	c map[net.Conn]bool
}{
	// This map is used as a set. Values are not consumed.
	c: make(map[net.Conn]bool),
}

// New returns a SunRPC server configured to listen on the given listener
func New(l net.Listener) *SunRPC {
	srv := &SunRPC{
		server:   rpc.NewServer(),
		listener: l,
		stop:     make(chan bool, 1),
	}

	programsList = []sunrpc.Program{
		newGfHandshake(),
		newGfDump(),
		pmap.NewGfPortmap(),
	}

	for _, p := range plugins.PluginsList {
		rpcProcs := p.SunRpcProgram()
		if rpcProcs != nil {
			programsList = append(programsList, rpcProcs)
			log.WithField("plugin", reflect.TypeOf(p)).Debug("loaded sunrpc procedures from plugin")
		}
	}

	port := getPortFromListener(srv.listener)

	for _, prog := range programsList {
		err := registerProgram(srv.server, prog, port)
		if err != nil {
			log.WithError(err).WithField("program", prog.Name()).Error("could not register SunRPC program")
			return nil
		}
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
		}
	}()

	log.WithField("ip:port", s.listener.Addr().String()).Info("started GlusterD SunRPC server")
	for {
		select {
		case <-s.stop:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			log.WithError(err).Error("failed to accept incoming connection")
			continue
		}

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
	close(s.stop)
	log.Info("Stopped GlusterD SunRPC server")
	// TODO: Gracefully stop the server
}
