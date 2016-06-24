// Package commands defines the command interfaces that need to be implemented by the GlusterD commands
package commands

import (
	"github.com/gluster/glusterd2/commands/peers"
	"github.com/gluster/glusterd2/commands/version"
	"github.com/gluster/glusterd2/commands/volumes"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
)

// Command is the interface that needs to be implemented by the GlusterD commands
type Command interface {
	// Routes should return a table of REST API endpoints and handlers for the command
	Routes() rest.Routes
	// Txns will setup the transaction details for the command
	Txns() *transaction.Txns
}

// Commands is a list of commands available
var Commands = []Command{
	&versioncommands.Command{},
	&volumecommands.Command{},
	&peercommands.Command{},
}
