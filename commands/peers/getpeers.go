package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	if peers, err := peer.GetPeersF(); err != nil {
		rest.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		rest.SendHTTPResponse(w, http.StatusOK, peers)
	}
}
