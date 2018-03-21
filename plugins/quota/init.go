package quota

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
)

const name = "quota"

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return name
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "QuotaList",
			Method:      "GET",
			Pattern:     "/quota/{volname}/limit",
			Version:     1,
			HandlerFunc: quotaListHandler},
		route.Route{
			Name:        "QuotaLimit",
			Method:      "POST",
			Pattern:     "/quota/{volname}/limit",
			Version:     1,
			HandlerFunc: quotaLimitHandler},
		route.Route{
			Name:        "QuotaRemove",
			Method:      "DELETE",
			Pattern:     "/quota/{volname}/limit",
			Version:     1,
			HandlerFunc: quotaRemoveHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	return
}
