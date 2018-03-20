// Package peercommands implements the peer management commands
package peercommands

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "GetPeer",
			Method:       "GET",
			Pattern:      "/peers/{peerid}",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.PeerGetResp)(nil)),
			HandlerFunc:  getPeerHandler,
		},
		route.Route{
			Name:         "GetPeers",
			Method:       "GET",
			Pattern:      "/peers",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.PeerListResp)(nil)),
			HandlerFunc:  getPeersHandler,
		},
		route.Route{
			Name:        "DeletePeer",
			Method:      "DELETE",
			Pattern:     "/peers/{peerid}",
			Version:     1,
			HandlerFunc: deletePeerHandler,
		},
		route.Route{
			Name:         "AddPeer",
			Method:       "POST",
			Pattern:      "/peers",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.PeerAddReq)(nil)),
			ResponseType: utils.GetTypeString((*api.PeerAddResp)(nil)),
			HandlerFunc:  addPeerHandler,
		},
		route.Route{
			Name:         "EditPeer",
			Method:       "POST",
			Pattern:      "/peers/{peerid}",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.PeerEditReq)(nil)),
			ResponseType: utils.GetTypeString((*api.PeerEditResp)(nil)),
			HandlerFunc:  editPeer,
		},
	}
}

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
	registerPeerEditStepFuncs()
}
