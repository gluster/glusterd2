package heketi

import (
	"github.com/gluster/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/transaction"
	"github.com/prashanthpai/sunrpc"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "heketi"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "HeketiDeviceAdd",
			Method:      "POST",
			Pattern:     "/heketi/{nodeid}/{devicename}/add",
			Version:     1,
			HandlerFunc: heketiDeviceAddHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnHeketiPrepareDevice, "heketi-prepare-device.Commit")
	transaction.RegisterStepFunc(txnHeketiCreateBrick, "heketi-create-brick.Commit")
}
