package peercommands

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/utils"
)

const (
	maxMetadataSizeLimit = 4096
)

func addPeerHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req api.PeerAddReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	if req.MetadataSize() > maxMetadataSizeLimit {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrMetadataSizeOutOfBounds)
		return
	}

	for key := range req.Metadata {
		if strings.HasPrefix(key, "_") {
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrRestrictedKeyFound)
			return
		}
	}

	if len(req.Addresses) < 1 {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrNoHostnamesPresent)
		return
	}
	logger.WithField("addresses", req.Addresses).Debug("received request to add new peer with given addresses")

	p, _ := peer.GetPeerByAddrs(req.Addresses)
	if p != nil {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, fmt.Sprintf("Peer exists with given addresses (ID: %s)", p.ID.String()))
		return
	}

	// A peer can have multiple addresses. For now, we use only the first
	// address present in the req.Addresses list.
	remotePeerAddress, err := utils.FormRemotePeerAddress(req.Addresses[0])
	if err != nil {
		logger.WithError(err).WithField("address", req.Addresses[0]).Error("failed to parse peer address")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "failed to parse remote address")
		return
	}

	// TODO: Try all addresses till the first one connects
	client, err := getPeerServiceClient(remotePeerAddress)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	defer client.conn.Close()
	logger = logger.WithField("peer", remotePeerAddress)

	newconfig := &StoreConfig{store.Store.Endpoints()}
	logger.WithField("endpoints", newconfig.Endpoints).Debug("asking new peer to join cluster with given endpoints")

	// Ask the peer to join the cluster
	rsp, err := client.JoinCluster(newconfig)
	if err != nil {
		logger.WithError(err).Error("sending Join request failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to send join cluster request")
		return
	} else if Error(rsp.Err) != ErrNone {
		err = Error(rsp.Err)
		logger.WithError(err).Error("join request failed")
		if rsp.Err == int32(ErrAnotherReqInProgress) {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	logger = logger.WithField("peerid", rsp.PeerID)
	logger.Info("new peer joined our cluster")

	// Get the new peer information to reply back with
	newpeer, err := peer.GetPeer(rsp.PeerID)
	if err != nil {
		logger.WithError(err).Error("failed to get peer information")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "new peer was added, but could not find peer in store. Try again later.")
		return
	}

	if req.Zone != "" {
		newpeer.Metadata["_zone"] = req.Zone
	}

	for key, value := range req.Metadata {
		newpeer.Metadata[key] = value
	}

	//check if remotePeerAddress already present
	found := utils.StringInSlice(remotePeerAddress, newpeer.PeerAddresses)
	//if not found prepend the remotePeerAddress to peer details
	if !found {
		newpeer.PeerAddresses = append([]string{remotePeerAddress}, newpeer.PeerAddresses...)
	} else {
		index := sort.StringSlice(newpeer.PeerAddresses).Search(remotePeerAddress)
		//swap peer  address to index 0
		sort.StringSlice(newpeer.PeerAddresses).Swap(0, index)
	}

	err = peer.AddOrUpdatePeer(newpeer)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Fail to add metadata to peer")
	}

	resp := createPeerAddResp(newpeer)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)

	// Save updated store endpoints for restarts
	store.Store.UpdateEndpoints()

	events.Broadcast(newPeerEvent(eventPeerAdded, newpeer))
}

func createPeerAddResp(p *peer.Peer) *api.PeerAddResp {
	return &api.PeerAddResp{
		ID:              p.ID,
		Name:            p.Name,
		PeerAddresses:   p.PeerAddresses,
		ClientAddresses: p.ClientAddresses,
		Metadata:        p.Metadata,
	}
}
