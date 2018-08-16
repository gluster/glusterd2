package peercommands

import (
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/servers/peerrpc"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/pborman/uuid"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var mutex = &utils.MutexWithTry{}

// PeerService implements the PeerService gRPC service
type PeerService int

func init() {
	peerrpc.Register(new(PeerService))
}

// RegisterService registers a service
func (p *PeerService) RegisterService(s *grpc.Server) {
	RegisterPeerServiceServer(s, p)
}

// Join makes the peer join the cluster of the requester
func (p *PeerService) Join(ctx context.Context, req *JoinReq) (*JoinRsp, error) {
	logger := log.WithFields(log.Fields{
		"remotepeer":    req.PeerID,
		"remotecluster": req.ClusterID})

	if mutex.TryLock() {
		defer mutex.Unlock()
	} else {
		logger.Info("rejecting join request, already processing another join/leave request")
		return &JoinRsp{"", int32(ErrAnotherReqInProgress)}, nil
	}

	logger.Info("handling new incoming join cluster request")

	// Handling a Join request happens as follows,
	// 	- TODO: Ensure no ongoing operations (transactions/other peer requests) are happening
	//      - Check if peer is part of another cluster
	// 	- Check if the peer has volumes
	//	- Reconfigure the store with received configuration
	// 	- Return your ID

	// TODO: Ensure no other operations are happening

	peers, err := peer.GetPeersF()
	if err != nil {
		logger.WithError(err).Error("failed to connect to store")
		return &JoinRsp{"", int32(ErrFailedToConnectToStore)}, nil
	}
	if len(peers) != 1 {
		logger.Info("rejecting join, already part of a cluster")
		return &JoinRsp{"", int32(ErrAnotherCluster)}, nil
	}

	volumes, err := volume.GetVolumes()
	if err != nil {
		logger.WithError(err).Error("failed to connect to store")
		return &JoinRsp{"", int32(ErrFailedToConnectToStore)}, nil
	}

	if len(volumes) != 0 {
		logger.Info("rejecting join, we already have volumes")
		return &JoinRsp{"", int32(ErrAnotherCluster)}, nil
	}

	logger.Debug("all checks passed, joining new cluster")

	// Update the cluster ID: This will set the global variable
	// gdctx.MyClusterID which will be used during store reconfiguration.
	// If reconfiguring store fails, restore the old cluster ID.
	success := false
	defer func(oldClusterID string) {
		if !success {
			gdctx.UpdateClusterID(oldClusterID)
		}
	}(gdctx.MyClusterID.String())
	if err := gdctx.UpdateClusterID(req.ClusterID); err != nil {
		return &JoinRsp{"", int32(ErrClusterIDUpdateFailed)}, nil
	}

	if err := ReconfigureStore(req.Config); err != nil {
		logger.WithError(err).Error("reconfigure store failed, failed to join new cluster")
		return &JoinRsp{"", int32(ErrStoreReconfigFailed)}, nil
	}
	success = true
	logger.Debug("reconfigured store to join new cluster")

	logger.Info("joined new cluster")
	return &JoinRsp{gdctx.MyUUID.String(), int32(ErrNone)}, nil
}

// Leave makes the peer leave its current cluster, and restart as a single node cluster
func (p *PeerService) Leave(ctx context.Context, req *LeaveReq) (*LeaveRsp, error) {
	logger := log.WithField("remotepeer", req.PeerID)

	if mutex.TryLock() {
		defer mutex.Unlock()
	} else {
		logger.Info("rejecting leave request, already processing another join/leave request")
		return &LeaveRsp{int32(ErrAnotherReqInProgress)}, nil
	}

	logger.Info("handling incoming leave cluster request")

	// Leaving a cluster happens in the following steps,
	// 	- TODO: Ensure no ongoing operations (transactions/other peer requests)
	// 	are happening
	// 	- Check if the request came from a known peer
	// 	- TODO: Check if you can leave the cluster
	// 	- Reconfigure the store with you defaults

	// TODO: Ensure no other operations are happening

	if p, err := peer.GetPeer(req.PeerID); err != nil {
		logger.Info("could not verify peer")
		return &LeaveRsp{int32(ErrUnknownPeer)}, nil
	} else if p == nil {
		logger.Info("rejecting leave, request received from unknown peer")
		return &LeaveRsp{int32(ErrUnknownPeer)}, nil
	}
	logger.Debug("request received from known peer")

	// TODO: Check if you can leave the cluster
	// The peer sending the leave request should have done the check, but check
	// again shouldn't hurt

	logger.Debug("all checks passed, leaving cluster")

	// Reset the cluster ID: This will reset the global variable
	// gdctx.MyClusterID which will be used during store reconfiguration.
	// If reconfiguring store fails, restore the old cluster ID.
	success := false
	defer func(oldClusterID string) {
		if !success {
			gdctx.UpdateClusterID(oldClusterID)
		}
	}(gdctx.MyClusterID.String())
	if err := gdctx.UpdateClusterID(uuid.New()); err != nil {
		return &LeaveRsp{int32(ErrClusterIDUpdateFailed)}, nil
	}

	logger.Debug("reconfiguring store with defaults")
	if err := ReconfigureStore(&StoreConfig{store.NewConfig().Endpoints}); err != nil {
		logger.WithError(err).Warn("failed to reconfigure store with defaults")
		// XXX: We should probably keep retrying here?
	}
	success = true
	return &LeaveRsp{int32(ErrNone)}, nil
}

// ReconfigureStore reconfigures the store with the given store config, if no
// store config is given uses the default
func ReconfigureStore(c *StoreConfig) error {

	// Destroy the current store first
	log.Debug("destroying current store")

	// Stop events framework
	events.Stop()

	// do not delete cluster namespace if this is not a loner node
	var deleteNamespace bool
	peers, err := peer.GetPeers()
	if err != nil {
		log.WithError(err).Error("failed to list peers during store reconfigure")
		return err
	}
	if len(peers) == 0 {
		// the peer entry for this node was removed from the store
		// by the node which received the peer removal request
		deleteNamespace = true
	}

	store.Destroy(deleteNamespace)

	// Restart the store with received configuration
	cfg := store.GetConfig()
	cfg.Endpoints = c.Endpoints

	if err := store.Init(cfg); err != nil {
		log.WithError(err).WithField("endpoints", cfg.Endpoints).Error("failed to restart store with new endpoints")
		// Restart store with default config
		defer restartDefaultStore(false, deleteNamespace)
		return err
	}
	log.WithField("endpoints", cfg.Endpoints).Debug("store restarted with new endpoints")

	// Save the new config if you successfully start the new store
	if err := cfg.Save(); err != nil {
		log.WithError(err).Error("failed to save new store configs")
		// Destroy newly started store and restart with default config
		defer restartDefaultStore(true, deleteNamespace)
		return err
	}
	log.Debug("saved new store config")

	// Add yourself to the peer list in the new store/cluster
	if err := peer.AddSelfDetails(); err != nil {
		log.WithError(err).Error("failed to add self to peer list")
		// Destroy newly started store and restart with default config
		defer restartDefaultStore(true, deleteNamespace)
		return err
	}
	log.Debug("added details of self to store")

	// Now that new store is up, start events framework
	events.Start()

	return nil
}

func restartDefaultStore(destroy bool, deleteNamespace bool) {
	if destroy {
		store.Destroy(deleteNamespace)
	}
	store.Init(nil)
	peer.AddSelfDetails()
	events.StartGlobal()
}
