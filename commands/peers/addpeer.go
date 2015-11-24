package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"

	"github.com/pborman/uuid"
)

type peerAddRequest struct {
	Addresses []string `json:"addresses"`
	Name      string   `json:"name,omitempty"`
}

func addPeerHandler(w http.ResponseWriter, r *http.Request) {
	var req peerAddRequest

	if e := utils.GetJSONFromRequest(r, &req); e != nil {
		utils.SendHTTPError(w, http.StatusBadRequest, e.Error())
		return
	}

	if len(req.Addresses) < 1 {
		utils.SendHTTPError(w, http.StatusBadRequest, errors.ErrNoHostnamesPresent.Error())
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

	if e := peer.AddOrUpdatePeer(p); e != nil {
		utils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	utils.SendHTTPResponse(w, http.StatusOK, nil)

}
