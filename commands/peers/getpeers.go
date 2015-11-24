package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	if peers, err := peer.GetPeers(); err != nil {
		utils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		utils.SendHTTPResponse(w, http.StatusOK, peers)
	}
}
