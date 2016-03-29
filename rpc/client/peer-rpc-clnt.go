package client

import (
	"fmt"
	"net"
	"net/rpc"

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
	//TODO : port 9876 is hardcoded for now, can be made configurable
	remoteAddress := fmt.Sprintf("%s:%s", p.Name, "9876")
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
	//TODO : port 9876 is hardcoded for now, can be made configurable
	remoteAddress := fmt.Sprintf("%s:%s", name, "9876")
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

// ConfigureETCDEnv function is a rpc server call for exporting and storing etcd
// environment variable
func ConfigureETCDEnv(p *peer.ETCDEnvConfig) (*services.RPCPeerGenericResp, error) {
	args := &services.RPCEtcdEnvReq{PeerName: new(string), Name: new(string), InitialCluster: new(string), ClusterState: new(string)}
	*args.PeerName = p.PeerName
	*args.Name = p.Name
	*args.InitialCluster = p.InitialCluster
	*args.ClusterState = p.ClusterState

	rsp := new(services.RPCPeerGenericResp)

	remoteAddress := fmt.Sprintf("%s:%s", p.PeerName, "9876")
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

	e = client.Call("PeerService.ExportAndStoreETCDEnv", args, rsp)
	if e != nil {
		log.Error("Failed to execute PeerService.ExportAndStoreEtcdEnv() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = &opRet
		rsp.OpError = &opError
		return rsp, e
	}
	return rsp, nil
}
