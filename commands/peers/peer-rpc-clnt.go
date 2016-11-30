package peercommands

import (
	"fmt"

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
func ValidateAddPeer(args *PeerAddReq) (*PeerAddResp, error) {
	remoteAddress := fmt.Sprintf("%s:%s", args.Address, config.GetString("rpcport"))
	rpcConn, e := grpc.Dial(remoteAddress, grpc.WithInsecure())
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"remote": remoteAddress,
		}).Error("failed to grpc.Dial remote")
		rsp := &PeerAddResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	defer rpcConn.Close()

	client := NewPeerServiceClient(rpcConn)

	rsp, e := client.ValidateAdd(netctx.TODO(), args)
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"rpc":    "PeerService.ValidateAdd",
			"remote": remoteAddress,
		}).Error("failed RPC call")
		rsp := &PeerAddResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	return rsp, nil
}

// ValidateDeletePeer is the validation function for DeletePeer to invoke the rpc
// server call
func ValidateDeletePeer(id string, ip string) (*PeerGenericResp, error) {
	args := &PeerDeleteReq{ID: id}

	remoteAddress := fmt.Sprintf("%s:%s", ip, config.GetString("rpcport"))
	rpcConn, e := grpc.Dial(remoteAddress, grpc.WithInsecure())
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"remote": remoteAddress,
		}).Error("failed to grpc.Dial remote")
		rsp := &PeerGenericResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	defer rpcConn.Close()

	client := NewPeerServiceClient(rpcConn)

	rsp, e := client.ValidateDelete(netctx.TODO(), args)
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"rpc":    "PeerService.ValidateDelete",
			"remote": remoteAddress,
		}).Error("failed RPC call")
		rsp := &PeerGenericResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	return rsp, nil
}

// ConfigureRemoteETCD will reconfigure etcd server on remote node to either
// join or remove itself from an etcd cluster.
func ConfigureRemoteETCD(remotePeer string, args *EtcdConfigReq) (*PeerGenericResp, error) {

	remoteAddress := fmt.Sprintf("%s:%s", remotePeer, config.GetString("rpcport"))
	rpcConn, e := grpc.Dial(remoteAddress, grpc.WithInsecure())
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"remote": remoteAddress,
		}).Error("failed to grpc.Dial remote")
		rsp := &PeerGenericResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	defer rpcConn.Close()

	client := NewPeerServiceClient(rpcConn)

	rsp, e := client.ExportAndStoreETCDConfig(netctx.TODO(), args)
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"rpc":    "PeerService.ExportAndStoreETCDConfig",
			"remote": remoteAddress,
		}).Error("failed RPC call")
		rsp := &PeerGenericResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	return rsp, nil
}
