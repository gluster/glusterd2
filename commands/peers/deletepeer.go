package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"

	"github.com/gorilla/mux"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		rest.SendHTTPError(w, http.StatusBadRequest, "peerid not present in the request")
		return
	}

	if !peer.Exists(id) {
		rest.SendHTTPError(w, http.StatusNotFound, "peer not found in cluster")
		return
	}

	if e := peer.DeletePeer(id); e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
	} else {
		rest.SendHTTPResponse(w, http.StatusNoContent, nil)
	}
}
