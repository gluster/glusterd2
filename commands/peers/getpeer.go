package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/context"

	"github.com/gorilla/mux"
)

func getPeer(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		client.sendResponse(w, -1, http.StatusBadRequest, "peerid not present in request", http.StatusBadRequest, nil)
		return
	}

	if peer, err := context.Store.GetPeer(id); err != nil {
		client.SendResponse(w, -1, http.StatusNotFound, err.Error(), http.StatusNotFound, "")
	} else {
		client.SendResponse(w, 0, 0, "", http.StatusOK, peer)
	}
}
