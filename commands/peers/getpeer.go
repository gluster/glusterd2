package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"

	"github.com/gorilla/mux"
)

func getPeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "peerid not present in request")
		return
	}

	if peer, err := peer.GetPeerF(id); err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusOK, peer)
	}
}
