package peercommands

import (
	log "github.com/Sirupsen/logrus"
	netctx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	opRet   int32
	opError string
)

// ValidateAddPeer is the validation function for AddPeer to invoke the rpc
// server call
func ValidateAddPeer(remoteAddress string, args *PeerAddReq) (*PeerAddResp, error) {
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
func ValidateDeletePeer(remoteAddress string, id string) (*PeerGenericResp, error) {
	args := &PeerDeleteReq{ID: id}

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
func ConfigureRemoteETCD(remoteAddress string, args *EtcdConfigReq) (*PeerGenericResp, error) {

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
