package georeplication

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
	return "georeplication"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "GeoreplicationCreate",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}",
			Version:     1,
			HandlerFunc: georepCreateHandler},
		route.Route{
			Name:        "GeoreplicationStart",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/start",
			Version:     1,
			HandlerFunc: georepStartHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnGeorepCreate, "georeplication-create.Commit")
	transaction.RegisterStepFunc(txnGeorepStart, "georeplication-start.Commit")
}
