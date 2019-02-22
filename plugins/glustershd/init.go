package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/utils"
	glustershdapi "github.com/gluster/glusterd2/plugins/glustershd/api"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "glustershd"
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "SelfHealInfo",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/{opts}/heal-info",
			Version:      1,
			ResponseType: utils.GetTypeString(([]glustershdapi.BrickHealInfo)(nil)),
			HandlerFunc:  selfhealInfoHandler},
		route.Route{
			Name:         "SelfHealInfo2",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/heal-info",
			Version:      1,
			ResponseType: utils.GetTypeString(([]glustershdapi.BrickHealInfo)(nil)),
			HandlerFunc:  selfhealInfoHandler},
		route.Route{
			Name:        "SelfHeal",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/heal",
			Version:     1,
			HandlerFunc: selfHealHandler},
		route.Route{
			Name:        "Split-Brain-Operations",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/split-brain/{operation}",
			Version:     1,
			RequestType: utils.GetTypeString(([]glustershdapi.SplitBrainReq)(nil)),
			HandlerFunc: splitBrainOperationHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	oldtransaction.RegisterStepFunc(txnSelfHeal, "selfheal.Heal")
}
