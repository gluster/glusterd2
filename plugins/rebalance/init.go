package rebalance

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/prashanthpai/sunrpc"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "rebalance"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "RebalanceStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/rebalance/start",
			Version:     1,
			HandlerFunc: rebalanceStart},
		route.Route{
			Name:        "RebalanceStop",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/rebalance/stop",
			Version:     1,
			HandlerFunc: rebalanceStop},
		route.Route{
			Name:        "RebalanceStatus",
			Method:      "GET",
			Pattern:     "/volumes/{volname}/rebalance/status",
			Version:     1,
			HandlerFunc: rebalanceStatus},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	registerRebalanceStartStepFuncs()
	registerRebalanceStopStepFuncs()
}
