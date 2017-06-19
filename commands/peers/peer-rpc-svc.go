package peercommands

import (
	"fmt"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/servers/peerrpc"
	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
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

// ValidateAdd validates AddPeer operation at server side
func (p *PeerService) ValidateAdd(ctx context.Context, args *PeerAddReq) (*PeerAddResp, error) {
	var opRet int32
	var opError string
	uuid := gdctx.MyUUID.String()

	log.Debug("recieved Validateadd peer request")

	volumes, _ := volume.GetVolumes()
	if len(volumes) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s already has existing volumes", args.Name)
		log.Info("rejecting peer add, we already have volumes")
	}

	reply := &PeerAddResp{
		OpRet:   opRet,
		OpError: opError,
		UUID:    uuid,
	}
	return reply, nil
}

// ValidateDelete validates DeletePeer operation at server side
func (p *PeerService) ValidateDelete(ctx context.Context, args *PeerDeleteReq) (*PeerGenericResp, error) {
	resp := &PeerGenericResp{}
	// TODO : Validate if this guy has any volume configured where the brick(s) is
	// hosted in some other node, in that case the validation should fail

	return resp, nil
}

// ReconfigureStore reconfigures the store with the recieved store config
func (p *PeerService) ReconfigureStore(ctx context.Context, c *StoreConfig) (*PeerGenericResp, error) {
	resp := &PeerGenericResp{}

	log.WithField("endpoints", c.Endpoints).Debug("recieved new reconfigure store request with new endpoints")

	// Stop the store first
	log.Debug("destroying store")
	store.Destroy()

	cfg := store.NewConfig()
	cfg.Endpoints = c.Endpoints

	log.Debug("restarting store with new endpoints")
	if err := store.Init(cfg); err != nil {
		resp.OpRet = -1
		resp.OpError = fmt.Sprintf("failed to restart store: %s", err.Error())
		// Restart store with default config
		defer peer.AddSelfDetails()
		defer store.Init(nil)
		return resp, nil
	}
	log.Debug("store restarted with new endpoints")

	if err := cfg.Save(); err != nil {
		resp.OpRet = -1
		resp.OpError = fmt.Sprintf("failed to save new store config: %s", err.Error())
		// Destroy newly started store and restart with default config
		defer peer.AddSelfDetails()
		defer store.Init(nil)
		defer store.Destroy()
		return resp, nil
	}
	log.Debug("saved new store config")

	if err := peer.AddSelfDetails(); err != nil {
		resp.OpRet = -1
		resp.OpError = fmt.Sprintf("could not add self details into etcd: %s", err.Error())
		// Destroy newly started store and restart with default config
		defer peer.AddSelfDetails()
		defer store.Init(nil)
		defer store.Destroy()
	}
	log.Debug("added details of self to store")

	return resp, nil
}
