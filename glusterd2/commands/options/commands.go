// Package optionscommands implements the commands to get and set cluster level options
package optionscommands

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
			Name:        "SetClusterOptions",
			Method:      "POST",
			Pattern:     "/cluster/options",
			Version:     1,
			HandlerFunc: setClusterOptionsHandler,
		},
		route.Route{
			Name:        "GetClusterOptions",
			Method:      "GET",
			Pattern:     "/cluster/options",
			Version:     1,
			HandlerFunc: getClusterOptionsHandler,
		},
	}
}

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
	return
}
