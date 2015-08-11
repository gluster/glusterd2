// The command package defines the command interfaces that need to be implemented by the GlusterD commands
package commands

import (
	"github.com/gorilla/mux"
	"github.com/kshlm/glusterd2/commands/hello"
	"github.com/kshlm/glusterd2/commands/volume-create"
	"github.com/kshlm/glusterd2/commands/volume-delete"
	"github.com/kshlm/glusterd2/commands/volume-info"
	"github.com/kshlm/glusterd2/commands/volume-list"
)

type Command interface {
	// SetRoutes should setup the REST API endpoints and handlers for the command
	SetRoutes(r *mux.Router) error
	// SetTransactionHandlers will setup the transaction handlers for the command
	//SetTransactionHandlers() error
}

var Commands = []Command{
	&hello.HelloCommand{},
	&volumecreate.VolumeCreateCommand{},
	&volumeinfo.VolumeInfoCommand{},
	&volumedelete.VolumeDeleteCommand{},
	&volumelist.VolumeListCommand{},
}
