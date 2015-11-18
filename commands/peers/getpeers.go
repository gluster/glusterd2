package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/peer"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	if peers, err := peer.GetPeers(); err != nil {
		client.SendResponse(w, -1, http.StatusNotFound, err.Error(), http.StatusNotFound, "")
	} else {
		client.SendResponse(w, 0, 0, "", http.StatusOK, peers)
	}
}
