package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"

	"github.com/gorilla/mux"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		utils.SendHTTPError(w, http.StatusBadRequest, "peerid not present in request")
		return
	}

	if peer, err := peer.GetPeer(id); err != nil {
		utils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		utils.SendHTTPResponse(w, http.StatusOK, peer)
	}
}
