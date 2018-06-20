// Package globalcommands implements the commands to get and set cluster level options
package globalcommands

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "SetGlobalOptions",
			Method:      "POST",
			Pattern:     "/cluster/options",
			Version:     1,
			HandlerFunc: setGlobalOptionsHandler,
		},
		route.Route{
			Name:        "GetGlobalOptions",
			Method:      "GET",
			Pattern:     "/cluster/options",
			Version:     1,
			HandlerFunc: getGlobalOptionsHandler,
		},
	}
}

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
	return
}
