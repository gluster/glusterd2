package georeplication

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
		route.Route{
			Name:        "GeoreplicationStop",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/stop",
			Version:     1,
			HandlerFunc: georepStopHandler},
		route.Route{
			Name:        "GeoreplicationDelete",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}",
			Version:     1,
			HandlerFunc: georepDeleteHandler},
		route.Route{
			Name:        "GeoreplicationPause",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/pause",
			Version:     1,
			HandlerFunc: georepPauseHandler},
		route.Route{
			Name:        "GeoreplicationResume",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/resume",
			Version:     1,
			HandlerFunc: georepResumeHandler},
		route.Route{
			Name:        "GeoreplicationStatus",
			Method:      "GET",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}",
			Version:     1,
			HandlerFunc: georepStatusHandler},
		route.Route{
			Name:        "GeoReplicationConfigGet",
			Method:      "GET",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/config",
			Version:     1,
			HandlerFunc: georepConfigGetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigSet",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/config",
			Version:     1,
			HandlerFunc: georepConfigSetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigReset",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{mastervolid}/{slavevolid}/config",
			Version:     1,
			HandlerFunc: georepConfigResetHandler,
		},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnGeorepCreate, "georeplication-create.Commit")
	transaction.RegisterStepFunc(txnGeorepStart, "georeplication-start.Commit")
	transaction.RegisterStepFunc(txnGeorepStop, "georeplication-stop.Commit")
	transaction.RegisterStepFunc(txnGeorepDelete, "georeplication-delete.Commit")
	transaction.RegisterStepFunc(txnGeorepPause, "georeplication-pause.Commit")
	transaction.RegisterStepFunc(txnGeorepResume, "georeplication-resume.Commit")
	transaction.RegisterStepFunc(txnGeorepStatus, "georeplication-status.Commit")
	transaction.RegisterStepFunc(txnGeorepConfigSet, "georeplication-configset.Commit")
	transaction.RegisterStepFunc(txnGeorepConfigFilegen, "georeplication-configfilegen.Commit")
}
