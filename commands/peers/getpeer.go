package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"

	"github.com/gorilla/mux"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "peerid not present in request", api.ErrCodeDefault)
		return
	}

	peer, err := peer.GetPeerF(id)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
	}

	resp := createPeerGetResp(peer)
	restutils.SendHTTPResponse(w, http.StatusOK, resp)
}

func createPeerGetResp(p *peer.Peer) *api.PeerGetResp {
	return &api.PeerGetResp{
		ID:        p.ID,
		Name:      p.Name,
		Addresses: p.Addresses,
	}
}
