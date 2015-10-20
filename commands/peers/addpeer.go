package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"

	"github.com/pborman/uuid"
)

type peerAddRequest struct {
	Addresses []string `json:"addresses"`
	Name      string   `json:"name,omitempty"`
}

func addPeer(w http.ResponseWriter, r *http.Request) {
	var req peerAddRequest

	if e := utils.GetJSONFromRequest(r, &req); e != nil {
		client.SendResponse(w, -1, http.StatusBadRequest, e.Error(), http.StatusBadRequest, nil)
		return
	}

	if len(req.Addresses) < 1 {
		client.SendResponse(w, -1, http.StatusBadRequest, "no hostnames present", http.StatusBadRequest, nil)
		return
	}

	if req.Name == "" {
		req.Name = req.Addresses[0]
	}

	//TODO: Do proper validation before initiating the add process

	//FIXME: In the correct add process, the peer being probed would add it's details to the store once it's been validated. The code below is just a temporary stand-in to show how the API's would work

	p := &peer.Peer{
		ID:        uuid.NewRandom(),
		Name:      req.Name,
		Addresses: req.Addresses,
	}

	if e := context.Store.AddOrUpdatePeer(p); e != nil {
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, nil)
		return
	}

	client.SendResponse(w, 0, 0, "", http.StatusOK, nil)

}
