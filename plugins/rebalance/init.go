package rebalance

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
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

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "RebalanceStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/rebalance/start",
			Version:     1,
			RequestType: utils.GetTypeString((*rebalanceapi.StartReq)(nil)),
			//			ResponseType: utils.GetTypeString((*rebalanceapi.RebalInfo)(nil)),
			HandlerFunc: rebalanceStartHandler},
		route.Route{
			Name:    "RebalanceStop",
			Method:  "POST",
			Pattern: "/volumes/{volname}/rebalance/stop",
			Version: 1,
			//			ResponseType: utils.GetTypeString((*rebalanceapi.RebalInfo)(nil)),
			HandlerFunc: rebalanceStopHandler},
		route.Route{
			Name:    "RebalanceStatus",
			Method:  "GET",
			Pattern: "/volumes/{volname}/rebalance",
			Version: 1,
			//			ResponseType: utils.GetTypeString((*rebalanceapi.RebalInfo)(nil)),
			HandlerFunc: rebalanceStatusHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnRebalanceStart, "rebalance-start")
	transaction.RegisterStepFunc(txnRebalanceStop, "rebalance-stop")
	transaction.RegisterStepFunc(txnRebalanceStatus, "rebalance-status")
	transaction.RegisterStepFunc(txnRebalanceStoreDetails, "rebalance-store")
}
