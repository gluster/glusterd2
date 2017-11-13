package peercommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/bin/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/peer"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	if peers, err := peer.GetPeersF(); err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusOK, peers)
	}
}
