package peercommands

import (
	"context"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	log "github.com/sirupsen/logrus"
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

// JoinCluster asks the remote peer to join the current cluster by reconfiguring the store with the given config
func (pc *peerSvcClnt) JoinCluster(conf *StoreConfig) (*JoinRsp, error) {
	args := &JoinReq{
		gdctx.MyUUID.String(),
		gdctx.MyClusterID.String(),
		conf,
	}
	rsp, err := pc.client.Join(context.TODO(), args)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"rpc":    "PeerService.Join",
			"remote": pc.address,
		}).Error("failed RPC call")
		return nil, err
	}
	return rsp, nil
}

// LeaveCluster asks the remote peer to leave the current cluster
func (pc *peerSvcClnt) LeaveCluster() (*LeaveRsp, error) {
	args := &LeaveReq{gdctx.MyUUID.String()}

	rsp, err := pc.client.Leave(context.TODO(), args)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"rpc":    "PeerService.Leave",
			"remote": pc.address,
		}).Error("failed RPC call")
		return nil, err
	}
	return rsp, nil
}
