package georeplication

import (
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/utils"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "georeplication"
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "GeoReplicationCreate",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCreateReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepCreateHandler},
		route.Route{
			Name:         "GeoReplicationStart",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/start",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepStartHandler},
		route.Route{
			Name:         "GeoReplicationStop",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/stop",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepStopHandler},
		route.Route{
			Name:        "GeoReplicationDelete",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{mastervolid}/{remotevolid}",
			Version:     1,
			HandlerFunc: georepDeleteHandler},
		route.Route{
			Name:         "GeoReplicationPause",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/pause",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepPauseHandler},
		route.Route{
			Name:         "GeoReplicationResume",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/resume",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepResumeHandler},
		route.Route{
			Name:         "GeoReplicationStatus",
			Method:       "GET",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepStatusHandler},
		route.Route{
			Name:         "GeoReplicationConfigGet",
			Method:       "GET",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/config",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepOption)(nil)),
			HandlerFunc:  georepConfigGetHandler,
		},
		route.Route{
			Name:         "GeoReplicationConfigSet",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/config",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepOption)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepOption)(nil)),
			HandlerFunc:  georepConfigSetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigReset",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{mastervolid}/{remotevolid}/config",
			Version:     1,
			HandlerFunc: georepConfigResetHandler,
		},
		route.Route{
			Name:         "GeoReplicationStatusList",
			Method:       "GET",
			Pattern:      "/geo-replication",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepSessionList)(nil)),
			HandlerFunc:  georepStatusListHandler},
		route.Route{
			Name:         "GeoReplicationSshKeyGenerate",
			Method:       "POST",
			Pattern:      "/ssh-key/{volname}/generate",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepSSHPublicKey)(nil)),
			HandlerFunc:  georepSSHKeyGenerateHandler},
		route.Route{
			Name:        "GeoReplicationSshKeyPush",
			Method:      "POST",
			Pattern:     "/ssh-key/{volname}/push",
			Version:     1,
			RequestType: utils.GetTypeString((*georepapi.GeorepSSHPublicKey)(nil)),
			HandlerFunc: georepSSHKeyPushHandler},
		route.Route{
			Name:         "GeoReplicationSshKeyGet",
			Method:       "GET",
			Pattern:      "/ssh-key/{volname}",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepSSHPublicKey)(nil)),
			HandlerFunc:  georepSSHKeyGetHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	oldtransaction.RegisterStepFunc(txnGeorepCreate, "georeplication-create.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepStart, "georeplication-start.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepStop, "georeplication-stop.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepDelete, "georeplication-delete.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepPause, "georeplication-pause.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepResume, "georeplication-resume.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepStatus, "georeplication-status.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepConfigSet, "georeplication-configset.Commit")
	oldtransaction.RegisterStepFunc(txnGeorepConfigFilegen, "georeplication-configfilegen.Commit")
	oldtransaction.RegisterStepFunc(txnSSHKeysGenerate, "georeplication-ssh-keygen.Commit")
	oldtransaction.RegisterStepFunc(txnSSHKeysPush, "georeplication-ssh-keypush.Commit")
}
