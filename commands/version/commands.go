// Package versioncommands implements the version command
package versioncommands

import (
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Txns returns command transaction steps. Required for the Command interface.
func (c *Command) Txns() *transaction.Txns {
	return &transaction.Txns{}

}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		rest.Route{
			Name:        "GetVersion",
			Method:      "GET",
			Pattern:     "/version",
			HandlerFunc: getVersionHandler,
		},
	}
}
