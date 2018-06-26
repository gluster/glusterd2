package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gorilla/mux"

	"github.com/pborman/uuid"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	id := mux.Vars(r)["peerid"]
	if uuid.Parse(id) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peer id passed")
		return
	}

	peer, err := peer.GetPeerF(id)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
	}

	resp := createPeerGetResp(peer)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createPeerGetResp(p *peer.Peer) *api.PeerGetResp {
	pid, online := store.Store.IsNodeAlive(p.ID)
	return &api.PeerGetResp{
		ID:              p.ID,
		Name:            p.Name,
		PeerAddresses:   p.PeerAddresses,
		ClientAddresses: p.ClientAddresses,
		Online:          online,
		PID:             pid,
		Metadata:        p.Metadata,
	}
}
