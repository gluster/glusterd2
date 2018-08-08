package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	id := mux.Vars(r)["peerid"]
	if uuid.Parse(id) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peer id passed")
		return
	}

	// Deleting a peer from the cluster happens as follows,
	// 	- Check if the peer is a member of the cluster
	// 	- Check if the peer can be removed
	//	- Delete the peer info from the store
	//	- Send the Leave request

	logger = logger.WithField("peerid", id)
	logger.Debug("received delete peer request")

	// Check whether the member exists
	p, err := peer.GetPeerF(id)
	if err != nil {
		logger.WithError(err).WithField("peerid", id).Error("Failed to get peer")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	// You cannot remove yourself
	if id == gdctx.MyUUID.String() {
		logger.Debug("request denied, received request to delete self from cluster")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "removing self is disallowed.")
		return
	}

	//Check if peer liveness key is present in store
	if _, alive := store.Store.IsNodeAlive(id); !alive {
		logger.Error("can not delete peer, peer is not alive")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "peer is not alive")
		return
	}

	// Check if any volumes exist with bricks on this peer
	if exists, err := bricksExist(id); err != nil {
		logger.WithError(err).Error("failed to check if bricks exist on peer")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "could not validate delete request")
		return
	} else if exists {
		logger.Debug("request denied, peer has bricks")
		restutils.SendHTTPError(ctx, w, http.StatusForbidden, "cannot delete peer, peer has bricks")
		return
	}

	remotePeerAddress, err := utils.FormRemotePeerAddress(p.PeerAddresses[0])
	if err != nil {
		logger.WithError(err).WithField("address", p.PeerAddresses[0]).Error("failed to parse peer address")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "failed to parse remote address")
		return
	}

	client, err := getPeerServiceClient(remotePeerAddress)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	defer client.conn.Close()

	// TODO: Need to do a better job of handling failures here. If this fails the
	// peer being removed still thinks it's a part of the cluster, and could
	// potentially still send commands to the cluster
	rsp, err := client.LeaveCluster()
	if err != nil {
		logger.WithError(err).Error("sending Leave request failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to send leave cluster request")
		return
	} else if Error(rsp.Err) != ErrNone {
		err = Error(rsp.Err)
		logger.WithError(err).Error("leave request failed")
		if rsp.Err == int32(ErrAnotherReqInProgress) {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	logger.Debug("peer left cluster")

	// Remove the peer details from the store
	if err := peer.DeletePeer(id); err != nil {
		logger.WithError(err).WithField("peer", id).Error("failed to remove peer from the store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)

	// Save updated store endpoints for restarts
	store.Store.UpdateEndpoints()

	events.Broadcast(newPeerEvent(eventPeerRemoved, p))
}

// bricksExist checks if the given peer has any bricks on it
// TODO: Move this to a more appropriate place
func bricksExist(id string) (bool, error) {
	pid := uuid.Parse(id)

	vols, err := volume.GetVolumes()
	if err != nil {
		return true, err
	}

	for _, v := range vols {
		for _, b := range v.GetBricks() {
			if uuid.Equal(pid, b.PeerID) {
				return true, nil
			}
		}
	}
	return false, nil
}
