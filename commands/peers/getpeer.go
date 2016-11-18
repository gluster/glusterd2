package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"

	"github.com/gorilla/mux"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		rest.SendHTTPError(w, http.StatusBadRequest, "peerid not present in request")
		return
	}

	if peer, err := peer.GetPeerF(id); err != nil {
		rest.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		rest.SendHTTPResponse(w, http.StatusOK, peer)
	}
}
