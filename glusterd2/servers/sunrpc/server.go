package sunrpc

import (
	"expvar"
	"io"
	"net"
	"net/rpc"
	"path"
	"strconv"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/pkg/sunrpc"

	"github.com/cockroachdb/cmux"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const gd2SocketFile = "glusterd2.socket"

var (
	// metrics
	clientCount = expvar.NewInt("sunrpc_clients_connected")
)

// SunRPC implements a suture service
type SunRPC struct {
	server        *rpc.Server
	tcpListener   net.Listener
	tcpStopCh     chan struct{}
	unixListener  net.Listener
	unixStopCh    chan struct{}
	notifyCloseCh chan io.ReadWriteCloser
}

var programsList []sunrpc.Program

// clientsList is global as it needs to be accessed by RPC procedures
// that notify connected clients.
var clientsList = struct {
	sync.RWMutex
	c map[net.Conn]struct{}
}{
	// This map is used as a set. Values are not consumed.
	c: make(map[net.Conn]struct{}),
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

	f := path.Join(config.GetString("rundir"), gd2SocketFile)
	uL, err := net.Listen("unix", f)
	if err != nil {
		// FIXME: Remove fatal and bubble up error to main()
		log.WithError(err).WithField("socket", gd2SocketFile).Fatal("failed to listen")
	}
	// This cleanup happens for process shutdown on SIGTERM/SIGINT but not on SIGKILL.
	uL.(*net.UnixListener).SetUnlinkOnClose(true)

	srv := &SunRPC{
		server:        rpc.NewServer(),
		tcpListener:   m.Match(sunrpc.CmuxMatcher()),
		unixListener:  uL,
		tcpStopCh:     make(chan struct{}),
		unixStopCh:    make(chan struct{}),
		notifyCloseCh: make(chan io.ReadWriteCloser, 10),
	}

	programsList = []sunrpc.Program{
		newGfHandshake(),
		newGfDump(),
		pmap.NewGfPortmap(),
	}

	port := getPortFromListener(srv.tcpListener)

	for _, prog := range programsList {
		err := registerProgram(srv.server, prog, port, false)
		if err != nil {
			log.WithError(err).WithField("program", prog.Name()).Error("could not register SunRPC program")
			return nil
		}
	}

	return srv
}

// pruneConn detects client disconnections and prunes clients list
func (s *SunRPC) pruneConn() {
	logger := log.WithField("server", "sunrpc")
	for rwc := range s.notifyCloseCh {
		conn := rwc.(net.Conn)
		logger.WithField("address", conn.RemoteAddr().String()).Info("client disconnected")

		clientsList.Lock()
		delete(clientsList.c, conn)
		clientsList.Unlock()

		clientCount.Add(-1)
	}
}

func (s *SunRPC) acceptLoop(stopCh chan struct{}, l net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()

	var ltype string
	switch l.(type) {
	case *net.UnixListener:
		ltype = "unix"
	default:
		ltype = "tcp"
	}
	logger := log.WithFields(log.Fields{
		"server":    "sunrpc",
		"transport": ltype})
	logger.WithField("address", l.Addr().String()).Info("started server")

	sessions := make([]rpc.ServerCodec, 50)
	for {
		select {
		case <-stopCh:
			logger.Debug("stopped accepting new connections")
			logger.Debug("closing client connections")
			for _, c := range sessions {
				if c != nil {
					c.Close()
				}
			}
			return
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			continue
		}
		logger.WithField("address", conn.RemoteAddr().String()).Info("client connected")
		clientCount.Add(1)
		clientsList.Lock()
		clientsList.c[conn] = struct{}{}
		clientsList.Unlock()

		session := sunrpc.NewServerCodec(conn, s.notifyCloseCh)
		go s.server.ServeCodec(session)
		sessions = append(sessions, session)
	}
}

// Serve will start accepting Sun RPC client connections on the listener
// provided.
func (s *SunRPC) Serve() {
	// FIXME: This goroutine leaks, the fix however makes code look complex.
	// We will need two separate servers once we decide that local daemons
	// only communicate over Unix sockets. Deferring this until then.
	go s.pruneConn()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.acceptLoop(s.tcpStopCh, s.tcpListener, wg)

	wg.Add(1)
	go s.acceptLoop(s.unixStopCh, s.unixListener, wg)

	wg.Wait()
}

// Stop stops the SunRPC server
func (s *SunRPC) Stop() {
	close(s.tcpStopCh)
	close(s.unixStopCh)

	// Close UDS listener; cmux should take care of the TCP one.
	s.unixListener.Close()
}
