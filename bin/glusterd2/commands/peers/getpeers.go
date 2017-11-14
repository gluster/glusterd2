package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/bin/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/bin/glusterd2/servers/rest/utils"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	if peers, err := peer.GetPeersF(); err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusOK, peers)
	}
}
