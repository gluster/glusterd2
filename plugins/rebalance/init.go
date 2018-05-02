package rebalance

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/sunrpc"
	"github.com/gluster/glusterd2/pkg/utils"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"
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
			Name:         "RebalanceStart",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/rebalance/start",
			Version:      1,
			RequestType:  utils.GetTypeString((*rebalanceapi.StartReq)(nil)),
			ResponseType: utils.GetTypeString((*rebalanceapi.RebalInfo)(nil)),
			HandlerFunc:  rebalanceStart},
		route.Route{
			Name:         "RebalanceStop",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/rebalance/stop",
			Version:      1,
			ResponseType: utils.GetTypeString((*rebalanceapi.RebalInfo)(nil)),
			HandlerFunc:  rebalanceStop},
		route.Route{
			Name:         "RebalanceStatus",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/rebalance",
			Version:      1,
			ResponseType: utils.GetTypeString((*rebalanceapi.RebalInfo)(nil)),
			HandlerFunc:  rebalanceStatus},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(startRebalance, "rebalance-start")
	transaction.RegisterStepFunc(storeRebalanceDetails, "rebalance.StoreVolume")
	transaction.RegisterStepFunc(stopRebalance, "rebalance-stop")
}
