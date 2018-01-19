package bitrot

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/prashanthpai/sunrpc"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "bitrot"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "BitrotEnable",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/bitrot/enable",
			Version:     1,
			HandlerFunc: bitrotEnableHandler},
		route.Route{
			Name:        "BitrotDisable",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/bitrot/disable",
			Version:     1,
			HandlerFunc: bitrotDisableHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnBitrotEnableDisable, "bitrot-enable.Commit")
	transaction.RegisterStepFunc(txnBitrotEnableDisable, "bitrot-disable.Commit")
	return
}
