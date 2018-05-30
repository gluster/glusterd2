// Package commands defines the command interfaces that need to be implemented by the GlusterD commands
package commands

import (
	"github.com/gluster/glusterd2/glusterd2/commands/global"
	"github.com/gluster/glusterd2/glusterd2/commands/peers"
	"github.com/gluster/glusterd2/glusterd2/commands/snapshot"
	"github.com/gluster/glusterd2/glusterd2/commands/version"
	"github.com/gluster/glusterd2/glusterd2/commands/volumes"
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
)

// Command is the interface that needs to be implemented by the GlusterD commands
type Command interface {
	// Routes should return a table of REST API endpoints and handlers for the command
	Routes() route.Routes
	// RegisterStepFuncs will register the transaction StepFuncs for the command
	RegisterStepFuncs()
}

// Commands is a list of commands available
var Commands = []Command{
	&versioncommands.Command{},
	&volumecommands.Command{},
	&snapshotcommands.Command{},
	&peercommands.Command{},
	&globalcommands.Command{},
}
