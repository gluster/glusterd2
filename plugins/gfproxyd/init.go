package gfproxyd

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/sunrpc"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "gfproxyd"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "GfproxydEnable",
			Method:      "POST",
			Pattern:     "/volumes/{name}/gfproxy/enable",
			Version:     1,
			HandlerFunc: gfproxydEnableHandler},
		route.Route{
			Name:        "GfproxydDisable",
			Method:      "POST",
			Pattern:     "/volumes/{name}/gfproxy/disable",
			Version:     1,
			HandlerFunc: gfproxydDisableHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnGfproxydStart, "gfproxyd-start.Commit")
	transaction.RegisterStepFunc(txnGfproxydStop, "gfproxyd-stop.Commit")
}
