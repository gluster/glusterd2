package services

import (
	"fmt"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/volume"
)

type PeerService int

var (
	opRet   int32
	opError string
)

func (p *PeerService) Validate(args *RPCPeerAddReq, reply *RPCPeerAddResp) error {
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
