package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gorilla/mux"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	id := mux.Vars(r)["peerid"]
	if id == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "peerid not present in request", api.ErrCodeDefault)
		return
	}

	peer, err := peer.GetPeerF(id)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
	}

	resp := createPeerGetResp(peer)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createPeerGetResp(p *peer.Peer) *api.PeerGetResp {
	return &api.PeerGetResp{
		ID:              p.ID,
		Name:            p.Name,
		PeerAddresses:   p.PeerAddresses,
		ClientAddresses: p.ClientAddresses,
		Online:          store.Store.IsNodeAlive(p.ID),
	}
}
