package sunrpc

import (
	"net"
	"net/rpc"
	"strconv"
	"reflect"

	"github.com/prashanthpai/sunrpc"
	"github.com/gluster/glusterd2/servers/sunrpc/program"
	"github.com/gluster/glusterd2/plugins"
	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

// SunRPC implements a suture service
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

	program.ProgramsList = []program.Program{
		newGfHandshake(),
		newGfPortmap(),
	}
	for _, p := range plugins.PluginsList{
		rpcProcs := p.SunRpcProcedures()
		if rpcProcs != nil{
			program.ProgramsList = append(program.ProgramsList, rpcProcs)
			log.WithField("plugin", reflect.TypeOf(p)).Debug("loaded sunrpc procedures from plugin")
		}
	}

	port := getPortFromListener(srv.listener)

	for _, prog := range program.ProgramsList {
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
	log.WithField("ip:port", s.listener.Addr().String()).Info("started GlusterD SunRPC server")
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
			log.WithError(err).Error("failed to accept incoming connection")
			continue
		}
		log.WithField("address", conn.RemoteAddr().String()).Info("glusterfs client connected")
		go s.server.ServeCodec(sunrpc.NewServerCodec(conn))
	}
}

// Stop stops the SunRPC server
func (s *SunRPC) Stop() {
	close(s.stop)
	log.Info("Stopped GlusterD SunRPC server")
	// TODO: Gracefully stop the server
}
