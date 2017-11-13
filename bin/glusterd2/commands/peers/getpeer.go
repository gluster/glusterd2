package peercommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/bin/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/peer"

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
