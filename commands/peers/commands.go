// Package peercommands implements the peer management commands
package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/rest"

	"github.com/gorilla/mux"
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
			HandlerFunc: getPeer,
		},
		rest.Route{
			Name:        "GetPeers",
			Method:      "GET",
			Pattern:     "/peers/",
			HandlerFunc: getPeers,
		},
		rest.Route{
			Name:        "DeletePeer",
			Method:      "DELETE",
			Pattern:     "/peers/{peerid}",
			HandlerFunc: deletePeer,
		},
		rest.Route{
			Name:        "AddPeer",
			Method:      "POST",
			Pattern:     "/peers/",
			HandlerFunc: addPeer,
		},
	}
}
