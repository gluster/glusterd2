package client

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/config"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rpc/services"

	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/pbrpc/pbcodec"
)

var (
	opRet   int32
	opError string
)

// ValidateAddPeer is the validation function for AddPeer to invoke the rpc
// server call
func ValidateAddPeer(p *peer.PeerAddRequest) (*services.RPCPeerAddResp, error) {
	args := &services.RPCPeerAddReq{Name: new(string), Addresses: p.Addresses}
	*args.Name = p.Name

	rsp := new(services.RPCPeerAddResp)
	remoteAddress := fmt.Sprintf("%s:%s", p.Name, config.RpcPort)
	rpcConn, e := net.Dial("tcp", remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}
	client := rpc.NewClientWithCodec(pbcodec.NewClientCodec(rpcConn))
	defer client.Close()

	e = client.Call("PeerService.ValidateAdd", args, rsp)
	if e != nil {
		log.Error("Failed to execute PeerService.ValidateAdd() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}
	return rsp, nil
}

// ValidateDeletePeer is the validation function for DeletePeer to invoke the rpc
// server call
func ValidateDeletePeer(id string, name string) (*services.RPCPeerGenericResp, error) {
	args := &services.RPCPeerDeleteReq{ID: new(string)}
	*args.ID = id

	rsp := new(services.RPCPeerGenericResp)
	remoteAddress := fmt.Sprintf("%s:%s", name, config.RpcPort)
	rpcConn, e := net.Dial("tcp", remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}
	client := rpc.NewClientWithCodec(pbcodec.NewClientCodec(rpcConn))
	defer client.Close()

	e = client.Call("PeerService.ValidateDelete", args, rsp)
	if e != nil {
		log.Error("Failed to execute PeerService.ValidateDelete() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}
	return rsp, nil
}

// ConfigureRemoteETCD function is a rpc server call for exporting and storing etcd
// environment variable & other configuration parameters
func ConfigureRemoteETCD(p *peer.ETCDConfig) (*services.RPCPeerGenericResp, error) {
	args := &services.RPCEtcdConfigReq{PeerName: new(string), Name: new(string), InitialCluster: new(string), ClusterState: new(string), Client: new(bool)}
	*args.PeerName = p.PeerName
	*args.Name = p.Name
	*args.InitialCluster = p.InitialCluster
	*args.ClusterState = p.ClusterState
	*args.Client = p.Client

	rsp := new(services.RPCPeerGenericResp)

	remoteAddress := fmt.Sprintf("%s:%s", p.PeerName, config.RpcPort)
	rpcConn, e := net.Dial("tcp", remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}

	client := rpc.NewClientWithCodec(pbcodec.NewClientCodec(rpcConn))
	defer client.Close()

	e = client.Call("PeerService.ExportAndStoreETCDConfig", args, rsp)
	if e != nil {
		log.Error("Failed to execute PeerService.ExportAndStoreEtcdConfig() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}
	return rsp, nil
}
