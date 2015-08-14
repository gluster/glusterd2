// The command package defines the command interfaces that need to be implemented by the GlusterD commands
package commands

import (
	"github.com/kshlm/glusterd2/commands/hello"
	"github.com/kshlm/glusterd2/commands/volume-create"
	"github.com/kshlm/glusterd2/commands/volume-delete"
	"github.com/kshlm/glusterd2/commands/volume-info"
	"github.com/kshlm/glusterd2/commands/volume-list"
	"github.com/kshlm/glusterd2/commands/volume-start"
	"github.com/kshlm/glusterd2/commands/volume-stop"
	"github.com/kshlm/glusterd2/rest"
)

type Command interface {
	// Routes should return a table of REST API endpoints and handlers for the command
	Routes() rest.Routes
	// SetTransactionHandlers will setup the transaction handlers for the command
	//SetTransactionHandlers() error
}

var Commands = []Command{
	&hello.HelloCommand{},
	&volumecreate.VolumeCreateCommand{},
	&volumeinfo.VolumeInfoCommand{},
	&volumedelete.VolumeDeleteCommand{},
	&volumelist.VolumeListCommand{},
	&volumestart.VolumeStartCommand{},
	&volumestop.VolumeStopCommand{},
}
