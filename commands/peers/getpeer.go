package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/peer"

	"github.com/gorilla/mux"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		client.SendResponse(w, -1, http.StatusBadRequest, "peerid not present in request", http.StatusBadRequest, nil)
		return
	}

	if peer, err := peer.GetPeer(id); err != nil {
		client.SendResponse(w, -1, http.StatusNotFound, err.Error(), http.StatusNotFound, "")
	} else {
		client.SendResponse(w, 0, 0, "", http.StatusOK, peer)
	}
}
