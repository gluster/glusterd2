package peercommands

import (
	"github.com/gluster/glusterd2/bin/glusterd2/gdctx"
	"github.com/gluster/glusterd2/bin/glusterd2/peer"
	"github.com/gluster/glusterd2/bin/glusterd2/servers/peerrpc"
	"github.com/gluster/glusterd2/bin/glusterd2/store"
	"github.com/gluster/glusterd2/bin/glusterd2/volume"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

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
	logger := log.WithField("remotepeer", req.PeerID)

	logger.Info("handling new incoming join cluster request")

	// Handling a Join request happens as follows,
	// 	- TODO: Ensure no ongoing operations (transactions/other peer requests) are happening
	// 	- TODO: Check is peer is part of another cluster
	// 	- Check if the peer has volumes
	//	- Reconfigure the store with received configuration
	// 	- Return your ID

	// TODO: Ensure no other operations are happening

	// TODO: Check if we are part of another cluster

	volumes, _ := volume.GetVolumes()
	if len(volumes) != 0 {
		logger.Info("rejecting join, we already have volumes")
		return &JoinRsp{"", int32(ErrAnotherCluster)}, nil
	}

	logger.Debug("all checks passed, joining new cluster")

	if err := ReconfigureStore(req.Config); err != nil {
		logger.WithError(err).Error("reconfigure store failed, failed to join new cluster")
		return &JoinRsp{"", int32(ErrStoreReconfigFailed)}, nil
	}
	logger.Debug("reconfigured store to join new cluster")

	logger.Info("joined new cluster")
	return &JoinRsp{gdctx.MyUUID.String(), int32(ErrNone)}, nil
}

// Leave makes the peer leave its current cluster, and restart as a single node cluster
func (p *PeerService) Leave(ctx context.Context, req *LeaveReq) (*LeaveRsp, error) {
	logger := log.WithField("remotepeer", req.PeerID)

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

	logger.Debug("reconfiguring store with defaults")
	if err := ReconfigureStore(&StoreConfig{store.NewConfig().Endpoints}); err != nil {
		logger.WithError(err).Warn("failed to reconfigure store with defaults")
		// XXX: We should probably keep retrying here?
	}
	return &LeaveRsp{int32(ErrNone)}, nil
}

// ReconfigureStore reconfigures the store with the given store config, if no
// store config is given uses the default
func ReconfigureStore(c *StoreConfig) error {

	// Destroy the current store first
	log.Debug("destroying current store")
	store.Destroy()
	// TODO: Also need to destroy any old files in localstatedir (eg. volfiles)

	// Restart the store with received configuration
	cfg := store.GetConfig()
	cfg.Endpoints = c.Endpoints

	if err := store.Init(cfg); err != nil {
		log.WithError(err).WithField("endpoints", cfg.Endpoints).Error("failed to restart store with new endpoints")
		// Restart store with default config
		defer peer.AddSelfDetails()
		defer store.Init(nil)
		return err
	}
	log.WithField("endpoints", cfg.Endpoints).Debug("store restarted with new endpoints")

	// Save the new config if you successfully start the new store
	if err := cfg.Save(); err != nil {
		log.WithError(err).Error("failed to save new store configs")
		// Destroy newly started store and restart with default config
		defer peer.AddSelfDetails()
		defer store.Init(nil)
		defer store.Destroy()
		return err
	}
	log.Debug("saved new store config")

	// Add yourself to the peer list in the new store/cluster
	if err := peer.AddSelfDetails(); err != nil {
		log.WithError(err).Error("failed to add self to peer list")
		// Destroy newly started store and restart with default config
		defer peer.AddSelfDetails()
		defer store.Init(nil)
		defer store.Destroy()
		return err
	}
	log.Debug("added details of self to store")

	return nil
}
