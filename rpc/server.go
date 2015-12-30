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

func (r *Connection) PeerAddRPCSvc(args *peer.PeerAddRequest, reply *RPCPeerAddResp) error {
	log.Debug("In PeerAdd")
	if context.MaxOpVersion < 40000 {
		*reply.OpRet = -1
		*reply.OpError = fmt.Sprintf("GlusterD instance running on %s is not compatible", args.Name)
	}
	peers, _ := peer.GetPeers()
	if len(peers) != 0 {
		*reply.OpRet = -1
		*reply.OpError = fmt.Sprintf("Peer %s is already part of another cluster", args.Name)
	}
	return nil
}

func StartListener() error {

	server := rpc.NewServer()
	server.Register(new(Connection))
	l, e := net.Listen("tcp", ":9876")
	if e != nil {
		log.WithField("error", e).Fatal("listener error")
		return e
	} else {
		log.Debug("listening on port 9876")
	}

	go func() {
		for {
			c, e := l.Accept()
			if e != nil {
				log.WithField("error", e).Info("Accept failed")
				continue
			}
			log.WithField("Connection", c.RemoteAddr()).Info("New incoming connection")
			go server.ServeCodec(pbcodec.NewServerCodec(c))
		}
	}()
	return nil
}
