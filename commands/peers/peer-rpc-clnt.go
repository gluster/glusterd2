package peercommands

import (
	"fmt"

	"github.com/gluster/glusterd2/peer"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
	netctx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	opRet   int32
	opError string
)

// ValidateAddPeer is the validation function for AddPeer to invoke the rpc
// server call
func ValidateAddPeer(p *peer.PeerAddRequest) (*PeerAddResp, error) {
	args := &PeerAddReq{Name: p.Name, Addresses: p.Addresses}

	rsp := new(PeerAddResp)
	remoteAddress := fmt.Sprintf("%s:%s", p.Name, config.GetString("rpcport"))
	rpcConn, e := grpc.Dial(remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = opRet
		rsp.OpError = opError
		return rsp, e
	}
	defer rpcConn.Close()

	client := NewPeerServiceClient(rpcConn)

	rsp, e = client.ValidateAdd(netctx.TODO(), args)
	if e != nil {
		log.Error("Failed to execute PeerService.ValidateAdd() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = opRet
		rsp.OpError = opError
		return rsp, e
	}
	return rsp, nil
}

// ValidateDeletePeer is the validation function for DeletePeer to invoke the rpc
// server call
func ValidateDeletePeer(id string, name string) (*PeerGenericResp, error) {
	args := &PeerDeleteReq{ID: id}

	rsp := new(PeerGenericResp)
	remoteAddress := fmt.Sprintf("%s:%s", name, config.GetString("rpcport"))
	rpcConn, e := grpc.Dial(remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = opRet
		rsp.OpError = opError
		return rsp, e
	}
	defer rpcConn.Close()

	client := NewPeerServiceClient(rpcConn)

	rsp, e = client.ValidateDelete(netctx.TODO(), args)
	if e != nil {
		log.Error("Failed to execute PeerService.ValidateDelete() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = opRet
		rsp.OpError = opError
		return rsp, e
	}
	return rsp, nil
}

// ConfigureRemoteETCD function is a rpc server call for exporting and storing etcd
// environment variable & other configuration parameters
func ConfigureRemoteETCD(p *peer.ETCDConfig) (*PeerGenericResp, error) {
	args := &EtcdConfigReq{
		PeerName:       p.PeerName,
		Name:           p.Name,
		InitialCluster: p.InitialCluster,
		ClusterState:   p.ClusterState,
		Client:         p.Client,
	}

	rsp := new(PeerGenericResp)

	remoteAddress := fmt.Sprintf("%s:%s", p.PeerName, config.GetString("rpcport"))
	rpcConn, e := grpc.Dial(remoteAddress)
	if e != nil {
		log.WithField("error", e).Error("net.Dial() call failed")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = opRet
		rsp.OpError = opError
		return rsp, e
	}
	defer rpcConn.Close()

	client := NewPeerServiceClient(rpcConn)

	rsp, e = client.ExportAndStoreETCDConfig(netctx.TODO(), args)
	if e != nil {
		log.Error("Failed to execute PeerService.ExportAndStoreEtcdConfig() rpc call")
		opRet = -1
		opError = e.Error()
		rsp.OpRet = opRet
		rsp.OpError = opError
		return rsp, e
	}
	return rsp, nil
}
