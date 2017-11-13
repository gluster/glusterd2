// Package peercommands implements the peer management commands
package peercommands

import (
	"github.com/gluster/glusterd2/bin/glusterd2/servers/rest/route"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "GetPeer",
			Method:      "GET",
			Pattern:     "/peers/{peerid}",
			Version:     1,
			HandlerFunc: getPeerHandler,
		},
		route.Route{
			Name:        "GetPeers",
			Method:      "GET",
			Pattern:     "/peers",
			Version:     1,
			HandlerFunc: getPeersHandler,
		},
		route.Route{
			Name:        "EtcdHealthPeer",
			Method:      "GET",
			Pattern:     "/peers/{peerid}/etcdhealth",
			Version:     1,
			HandlerFunc: peerEtcdHealthHandler,
		},
		route.Route{
			Name:        "EtcdStatusPeer",
			Method:      "GET",
			Pattern:     "/peers/{peerid}/etcdstatus",
			Version:     1,
			HandlerFunc: peerEtcdStatusHandler,
		},
		route.Route{
			Name:        "DeletePeer",
			Method:      "DELETE",
			Pattern:     "/peers/{peerid}",
			Version:     1,
			HandlerFunc: deletePeerHandler,
		},
		route.Route{
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
