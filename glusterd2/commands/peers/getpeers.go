package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	peers, err := peer.GetPeersF()
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
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
