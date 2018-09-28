package daemon

import (
	"io"
	"net"
	"net/rpc"
	"sync"

	"github.com/gluster/glusterd2/pkg/sunrpc"
	log "github.com/sirupsen/logrus"
)

type connection struct {
	Transport *net.UnixConn
	RPC       *rpc.Client
}

var connectionsList = struct {
	sync.RWMutex
	c map[string]connection
}{
	c: make(map[string]connection),
}

// GetRPCClient returns *rpc.Client for the daemon. If a prior connection
// exists, it's returned. If not, a new RPC connection is created and
// returned.
func GetRPCClient(d Daemon) (*rpc.Client, error) {

	// FIXME: Reconnections happen only on-demand.

	connectionsList.RLock()
	dConn, ok := connectionsList.c[d.ID()]
	if ok {
		connectionsList.RUnlock()
		return dConn.RPC, nil
	}
	connectionsList.RUnlock()

	conn, err := net.DialUnix("unix", nil,
		&net.UnixAddr{Name: d.SocketFile(), Net: "unix"},
	)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"socket": d.SocketFile(), "name": d.Name(),
		}).Error("failed connecting to daemon")
		return nil, err
	}

	client := rpc.NewClientWithCodec(sunrpc.NewClientCodec(conn, notifyClose))
	dConn = connection{
		Transport: conn,
		RPC:       client,
	}

	connectionsList.Lock()
	connectionsList.c[d.ID()] = dConn
	connectionsList.Unlock()

	log.WithFields(log.Fields{
		"socket": d.SocketFile(), "name": d.Name(),
	}).Info("connected to daemon")

	return client, nil
}

var notifyClose = make(chan io.ReadWriteCloser, 10)

func init() {
	// Get notified on disconnections and prune the connections list.
	go func() {
		for rwc := range notifyClose {
			conn := rwc.(*net.UnixConn)
			log.WithField("socket", conn.RemoteAddr().String()).Info("daemon disconnected")
			connectionsList.Lock()
			for k, v := range connectionsList.c {
				if v.Transport == conn {
					delete(connectionsList.c, k)
				}
			}
			connectionsList.Unlock()
		}
	}()
}
