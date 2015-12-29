package rpc

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/peer"

	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/pbrpc/pbcodec"
)

func PeerAddRPCClnt(p *peer.PeerAddRequest) (*RPCResponse, error) {
	rsp := new(RPCResponse)
	remoteAddress := fmt.Sprintf("%s:%s", p.Name, "9876")
	rpcConn, e := net.Dial("tcp", remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		rsp.OpRet = -1
		rsp.OpError = e.Error()
		return rsp, e
	}
	client := rpc.NewClientWithCodec(pbcodec.NewClientCodec(rpcConn))
	defer client.Close()

	e = client.Call("Connection.PeerAddRPCSvc", p, rsp)
	if e != nil {
		log.Error("Failed to execute PeerAddRPCSvc() rpc call")
		rsp.OpRet = -1
		rsp.OpError = e.Error()
		return rsp, e
	}
	return rsp, nil
}
