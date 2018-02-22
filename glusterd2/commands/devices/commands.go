package devicecommands

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "DeviceAdd",
			Method:       "POST",
			Pattern:      "/peers/{peerid}/devices",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.AddDeviceReq)(nil)),
			ResponseType: utils.GetTypeString((*api.DeviceAddResp)(nil)),
			HandlerFunc:  deviceAddHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (c *Command) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnPrepareDevice, "prepare-device")
}
