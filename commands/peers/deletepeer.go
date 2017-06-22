package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {
	peerReq := mux.Vars(r)

	id := peerReq["peerid"]
	if id == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "peerid not present in the request")
		return
	}

	// Deleting a peer from the cluster happens as follows,
	// 	- Check if the peer is a member of the cluster
	// 	- Check if the peer can be removed
	//	- Send the Leave request
	//	- Delete the peer info from the store

	logger := log.WithField("peerid", id)
	logger.Debug("recieved delete peer request")

	// Check whether the member exists
	p, err := peer.GetPeerF(id)
	if err != nil {
		logger.WithError(err).Error("failed to get peer")
		restutils.SendHTTPError(w, http.StatusInternalServerError, "could not validate delete request")
		return
	} else if p == nil {
		logger.Debug("request denied, recieved request to remove unknown peer")
		restutils.SendHTTPError(w, http.StatusNotFound, "peer not found in cluster")
		return
	}

	// You cannot remove yourself
	if id == gdctx.MyUUID.String() {
		logger.Debug("request denied, recieved request to delete self from cluster")
		restutils.SendHTTPError(w, http.StatusBadRequest, "removing self is disallowed.")
		return
	}

	// Check if any volumes exist with bricks on this peer
	if exists, err := bricksExist(id); err != nil {
		logger.WithError(err).Error("failed to check if bricks exist on peer")
		restutils.SendHTTPError(w, http.StatusInternalServerError, "could not validate delete request")
		return
	} else if exists {
		logger.Debug("request denied, peer has bricks")
		restutils.SendHTTPError(w, http.StatusForbidden, "cannot delete peer, peer has bricks")
		return
	}

	remotePeerAddress, err := utils.FormRemotePeerAddress(p.Addresses[0])
	if err != nil {
		log.WithError(err).WithField("address", req.Addresses[0]).Error("failed to parse peer address")
		restutils.SendHTTPError(w, http.StatusBadRequest, "failed to parse remote address")
		return
	}

	client, err := getPeerServiceClient(remotePeerAddress)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer client.conn.Close()

	rsp, err := client.LeaveCluster()
	if err != nil {
		logger.WithError(err).Error("sending Leave request failed")
		restutils.SendHTTPError(w, http.StatusInternalServerError, "failed to send leave cluster request")
		return
	} else if Error(rsp.Err) != ErrNone {
		err = Error(rsp.Err)
		logger.WithError(err).Error("leave request failed")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	logger.Debug("peer left cluster")

	// Remove the peer details from the store
	if err := peer.DeletePeer(id); err != nil {
		log.WithError(err).WithField("peer", id).Error("failed to remove peer from the store")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusNoContent, nil)
	}

	// Save updated store endpoints for restarts
	store.Store.UpdateEndpoints()
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
		for _, b := range v.Bricks {
			if uuid.Equal(pid, b.NodeID) {
				return true, nil
			}
		}
	}
	return false, nil
}
