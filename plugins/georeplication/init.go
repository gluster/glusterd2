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
	return "geo-replication"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "GeoReplicationCreate",
			Method:      "PUT",
			Pattern:     "/geo-replication/{masterid}/{slaveid}",
			Version:     1,
			HandlerFunc: georepCreateHandler,
		},
		route.Route{
			Name:        "GeoReplicationUpdate",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}",
			Version:     1,
			HandlerFunc: georepCreateHandler,
		},
		route.Route{
			Name:        "GeoReplicationStart",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/start",
			Version:     1,
			HandlerFunc: georepStartHandler,
		},
		route.Route{
			Name:        "GeoReplicationPause",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/pause",
			Version:     1,
			HandlerFunc: georepPauseHandler,
		},
		route.Route{
			Name:        "GeoReplicationResume",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/resume",
			Version:     1,
			HandlerFunc: georepResumeHandler,
		},
		route.Route{
			Name:        "GeoReplicationStop",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/stop",
			Version:     1,
			HandlerFunc: georepStopHandler,
		},
		route.Route{
			Name:        "GeoReplicationDelete",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{masterid}/{slaveid}",
			Version:     1,
			HandlerFunc: georepDeleteHandler,
		},
		route.Route{
			Name:        "GeoReplicationStatus",
			Method:      "GET",
			Pattern:     "/geo-replication/{masterid}/{slaveid}",
			Version:     1,
			HandlerFunc: georepStatusHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigGet",
			Method:      "GET",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/config",
			Version:     1,
			HandlerFunc: georepConfigGetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigSet",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/config",
			Version:     1,
			HandlerFunc: georepConfigSetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigReset",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/config",
			Version:     1,
			HandlerFunc: georepConfigResetHandler,
		},
		route.Route{
			Name:        "GeoReplicationCheckpointSet",
			Method:      "POST",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/checkpoint",
			Version:     1,
			HandlerFunc: georepCheckpointSetHandler,
		},
		route.Route{
			Name:        "GeoReplicationCheckpointReset",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/checkpoint",
			Version:     1,
			HandlerFunc: georepCheckpointResetHandler,
		},
		route.Route{
			Name:        "GeoReplicationCheckpointGet",
			Method:      "GET",
			Pattern:     "/geo-replication/{masterid}/{slaveid}/checkpoint",
			Version:     1,
			HandlerFunc: georepCheckpointGetHandler,
		},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnGsyncdCreate, "georeplication-create.Commit")
	transaction.RegisterStepFunc(txnGsyncdStart, "georeplication-start.Commit")
	transaction.RegisterStepFunc(txnGsyncdPause, "georeplication-pause.Commit")
	transaction.RegisterStepFunc(txnGsyncdResume, "georeplication-resume.Commit")
	transaction.RegisterStepFunc(txnGsyncdStop, "georeplication-stop.Commit")
	// transaction.RegisterStepFunc(txnGsyncdDelete, "georeplication-delete.Commit")
	transaction.RegisterStepFunc(txnGsyncdStatus, "georeplication-status.Commit")
}
