package rpc

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/pbrpc/pbcodec"
)

var (
	opRet   int32
	opError string
)

func (r *Connection) PeerAddRPCSvc(args *RPCPeerAddReq, reply *RPCPeerAddResp) error {
	opRet = 0
	opError = ""
	if context.MaxOpVersion < 40000 {
		opRet = -1
		opError = fmt.Sprintf("GlusterD instance running on %s is not compatible", args.Name)
	}
	peers, _ := peer.GetPeers()
	if len(peers) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s is already part of another cluster", args.Name)
	}
	volumes, _ := volume.GetVolumes()
	if len(volumes) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s already has existing volumes", args.Name)
	}

	reply.OpRet = &opRet
	reply.OpError = &opError

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
