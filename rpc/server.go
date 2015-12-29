package rpc

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/peer"

	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/pbrpc/pbcodec"
)

func (r *Connection) PeerAddRPCSvc(args *peer.PeerAddRequest, reply *RPCResponse) error {
	log.Debug("In PeerAdd")
	if context.MaxOpVersion < 40000 {
		reply.OpRet = -1
		fmt.Sprintf(reply.OpError, "GlusterD instance on %s is not compatible", args.Name)
	}
	peers, _ := peer.GetPeers()
	if len(peers) != 0 {
		reply.OpRet = -1
		fmt.Sprintf(reply.OpError, "Peer %s is already part of another cluster", args.Name)
	}
	return nil
}

func RegisterServer() error {

	server := rpc.NewServer()
	server.Register(new(Connection))
	server.RegisterName("Connection", new(Connection))
	l, e := net.Listen("tcp", ":9876")
	if e != nil {
		log.WithField("error", e).Fatal("listen error")
		return e
	} else {
		log.Debug("listening on port 9876")
	}

	go func() {
		for {
			c, e := l.Accept()
			if e != nil {
				continue
			}
			log.WithField("Connection", c.RemoteAddr()).Info("New incoming connection")
			server.ServeCodec(pbcodec.NewServerCodec(c))
		}
	}()
	return nil
}
