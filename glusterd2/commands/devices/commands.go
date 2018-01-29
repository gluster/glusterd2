package devicecommands

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "DeviceAdd",
			Method:      "POST",
			Pattern:     "/devices",
			Version:     1,
			HandlerFunc: deviceAddHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (c *Command) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnPrepareDevice, "prepare-device.Commit")
}
