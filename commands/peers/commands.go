// Package peercommands implements the peer management commands
package peercommands

import (
	"github.com/gluster/glusterd2/rest"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		rest.Route{
			Name:        "GetPeer",
			Method:      "GET",
			Pattern:     "/peers/{peerid}",
			Version:     1,
			HandlerFunc: getPeerHandler,
		},
		rest.Route{
			Name:        "GetPeers",
			Method:      "GET",
			Pattern:     "/peers",
			Version:     1,
			HandlerFunc: getPeersHandler,
		},
		rest.Route{
			Name:        "DeletePeer",
			Method:      "DELETE",
			Pattern:     "/peers/{peerid}",
			Version:     1,
			HandlerFunc: deletePeerHandler,
		},
		rest.Route{
			Name:        "AddPeer",
			Method:      "POST",
			Pattern:     "/peers",
			Version:     1,
			HandlerFunc: addPeerHandler,
		},
	}
}

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
	return
}
