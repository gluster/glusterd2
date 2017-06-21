package peercommands

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	opRet   int32
	opError string
)

type peerSvcClnt struct { // this is not really a good name as it can be confused with PeerServiceClient, but there isn't anything better
	conn    *grpc.ClientConn
	client  PeerServiceClient
	address string
}

// getPeerServiceClient returns a PeerServiceClient for the given address and the underlying grpc.ClientConn
func getPeerServiceClient(address string) (*peerSvcClnt, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	clnt := NewPeerServiceClient(conn)

	return &peerSvcClnt{conn, clnt, address}, nil
}

// ValidateAddPeer is the validation function for AddPeer to invoke the rpc
// server call
func (pc *peerSvcClnt) ValidateAddPeer(args *PeerAddReq) (*PeerAddResp, error) {
	rsp, e := pc.client.ValidateAdd(context.TODO(), args)
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"rpc":    "PeerService.ValidateAdd",
			"remote": pc.address,
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
func (pc *peerSvcClnt) ValidateDeletePeer(id string) (*PeerGenericResp, error) {
	args := &PeerDeleteReq{ID: id}

	rsp, e := pc.client.ValidateDelete(context.TODO(), args)
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"rpc":    "PeerService.ValidateDelete",
			"remote": pc.address,
		}).Error("failed RPC call")
		rsp := &PeerGenericResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	return rsp, nil
}

// JoinCluster reconfigures the store of the newpeer to add it to the cluster
func (pc *peerSvcClnt) JoinCluster(args *StoreConfig) (*PeerGenericResp, error) {
	rsp, e := pc.client.ReconfigureStore(context.TODO(), args)
	if e != nil {
		log.WithFields(log.Fields{
			"error":  e,
			"rpc":    "PeerService.ReconfigureStore",
			"remote": pc.address,
		}).Error("failed RPC call")
		rsp := &PeerGenericResp{
			OpRet:   -1,
			OpError: e.Error(),
		}
		return rsp, e
	}
	return rsp, nil
}
