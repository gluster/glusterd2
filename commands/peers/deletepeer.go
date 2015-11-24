package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"

	"github.com/gorilla/mux"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		utils.SendHTTPError(w, http.StatusBadRequest, "peerid not present in the request")
		return
	}

	if !peer.Exists(id) {
		utils.SendHTTPError(w, http.StatusNotFound, "peer not found in cluster")
		return
	}

	if e := peer.DeletePeer(id); e != nil {
		utils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
	} else {
		utils.SendHTTPResponse(w, http.StatusNoContent, nil)
	}
}
