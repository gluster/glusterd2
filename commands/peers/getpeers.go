package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	peers, err := peer.GetPeersF()
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	}

	resp := createPeerListResp(peers)
	restutils.SendHTTPResponse(w, http.StatusOK, resp)
}

func createPeerListResp(peers []*peer.Peer) *api.PeerListResp {
	var resp api.PeerListResp

	for _, p := range peers {
		resp = append(resp, api.PeerGetResp{
			ID:        p.ID,
			Name:      p.Name,
			Addresses: p.Addresses,
		})
	}

	return &resp
}
