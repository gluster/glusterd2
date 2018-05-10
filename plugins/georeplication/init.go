package georeplication

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
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
			Name:         "GeoreplicationCreate",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCreateReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepCreateHandler},
		route.Route{
			Name:         "GeoreplicationStart",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/start",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepStartHandler},
		route.Route{
			Name:         "GeoreplicationStop",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/stop",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepStopHandler},
		route.Route{
			Name:        "GeoreplicationDelete",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{mastervolid}/{remotevolid}",
			Version:     1,
			HandlerFunc: georepDeleteHandler},
		route.Route{
			Name:         "GeoreplicationPause",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/pause",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepPauseHandler},
		route.Route{
			Name:         "GeoreplicationResume",
			Method:       "POST",
			Pattern:      "/geo-replication/{mastervolid}/{remotevolid}/resume",
			Version:      1,
			RequestType:  utils.GetTypeString((*georepapi.GeorepCommandsReq)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepResumeHandler},
		route.Route{
			Name:         "GeoreplicationStatus",
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
			RequestType:  utils.GetTypeString((*georepapi.GeorepOption)(nil)),
			ResponseType: utils.GetTypeString((*georepapi.GeorepOption)(nil)),
			HandlerFunc:  georepConfigGetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigSet",
			Method:      "POST",
			Pattern:     "/geo-replication/{mastervolid}/{remotevolid}/config",
			Version:     1,
			HandlerFunc: georepConfigSetHandler,
		},
		route.Route{
			Name:        "GeoReplicationConfigReset",
			Method:      "DELETE",
			Pattern:     "/geo-replication/{mastervolid}/{remotevolid}/config",
			Version:     1,
			HandlerFunc: georepConfigResetHandler,
		},
		route.Route{
			Name:         "GeoreplicationStatusList",
			Method:       "GET",
			Pattern:      "/geo-replication",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepSession)(nil)),
			HandlerFunc:  georepStatusListHandler},
		route.Route{
			Name:         "GeoreplicationSshKeyGenerate",
			Method:       "POST",
			Pattern:      "/ssh-key/{volname}/generate",
			Version:      1,
			ResponseType: utils.GetTypeString((*georepapi.GeorepSSHPublicKey)(nil)),
			HandlerFunc:  georepSSHKeyGenerateHandler},
		route.Route{
			Name:        "GeoreplicationSshKeyPush",
			Method:      "POST",
			Pattern:     "/ssh-key/{volname}/push",
			Version:     1,
			RequestType: utils.GetTypeString((*georepapi.GeorepSSHPublicKey)(nil)),
			HandlerFunc: georepSSHKeyPushHandler},
		route.Route{
			Name:         "GeoreplicationSshKeyGet",
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
	transaction.RegisterStepFunc(txnGeorepCreate, "georeplication-create.Commit")
	transaction.RegisterStepFunc(txnGeorepStart, "georeplication-start.Commit")
	transaction.RegisterStepFunc(txnGeorepStop, "georeplication-stop.Commit")
	transaction.RegisterStepFunc(txnGeorepDelete, "georeplication-delete.Commit")
	transaction.RegisterStepFunc(txnGeorepPause, "georeplication-pause.Commit")
	transaction.RegisterStepFunc(txnGeorepResume, "georeplication-resume.Commit")
	transaction.RegisterStepFunc(txnGeorepStatus, "georeplication-status.Commit")
	transaction.RegisterStepFunc(txnGeorepConfigSet, "georeplication-configset.Commit")
	transaction.RegisterStepFunc(txnGeorepConfigFilegen, "georeplication-configfilegen.Commit")
	transaction.RegisterStepFunc(txnSSHKeysGenerate, "georeplication-ssh-keygen.Commit")
	transaction.RegisterStepFunc(txnSSHKeysPush, "georeplication-ssh-keypush.Commit")
}
