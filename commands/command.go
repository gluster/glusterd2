// Package commands defines the command interfaces that need to be implemented by the GlusterD commands
package commands

import (
	"github.com/kshlm/glusterd2/commands/hello"
	"github.com/kshlm/glusterd2/commands/volume-add-brick"
	"github.com/kshlm/glusterd2/commands/volume-create"
	"github.com/kshlm/glusterd2/commands/volume-delete"
	"github.com/kshlm/glusterd2/commands/volume-info"
	"github.com/kshlm/glusterd2/commands/volume-list"
	"github.com/kshlm/glusterd2/commands/volume-remove-brick"
	"github.com/kshlm/glusterd2/commands/volume-start"
	"github.com/kshlm/glusterd2/commands/volume-stop"
	"github.com/kshlm/glusterd2/rest"
)

// Command is the interface that needs to be implemented by the GlusterD commands
type Command interface {
	// Routes should return a table of REST API endpoints and handlers for the command
	Routes() rest.Routes
	// SetTransactionHandlers will setup the transaction handlers for the command
	//SetTransactionHandlers() error
}

// Commands is a list of commands available
var Commands = []Command{
	&hello.Command{},
	&volumecreate.Command{},
	&volumeinfo.Command{},
	&volumedelete.Command{},
	&volumelist.Command{},
	&volumestart.Command{},
	&volumestop.Command{},
	&volumeaddbrick.Command{},
	&volumeremovebrick.Command{},
}
