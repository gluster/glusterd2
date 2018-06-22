package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
)

func getPeersHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	keys, keyFound := r.URL.Query()["key"]
	values, valueFound := r.URL.Query()["value"]
	filterParams := make(map[string]string)
	if keyFound {
		filterParams["key"] = keys[0]
	}
	if valueFound {
		filterParams["value"] = values[0]
	}
	peers, err := peer.GetPeersF(filterParams)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := createPeerListResp(peers)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createPeerListResp(peers []*peer.Peer) *api.PeerListResp {
	var resp api.PeerListResp

	for _, p := range peers {
		pid, online := store.Store.IsNodeAlive(p.ID)
		resp = append(resp, api.PeerGetResp{
			ID:              p.ID,
			Name:            p.Name,
			PeerAddresses:   p.PeerAddresses,
			ClientAddresses: p.ClientAddresses,
			Online:          online,
			PID:             pid,
			Metadata:        p.Metadata,
		})
	}

	return &resp
}
