package tracemgmt

import (
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/utils"
	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "tracemgmt"
}

// RestRoutes returns a list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "TraceEnable",
			Method:       "POST",
			Pattern:      "/tracemgmt",
			Version:      1,
			RequestType:  utils.GetTypeString((*tracemgmtapi.SetupTracingReq)(nil)),
			ResponseType: utils.GetTypeString((*tracemgmtapi.JaegerConfigInfo)(nil)),
			HandlerFunc:  tracingEnableHandler},
		route.Route{
			Name:         "TraceStatus",
			Method:       "GET",
			Pattern:      "/tracemgmt",
			Version:      1,
			ResponseType: utils.GetTypeString((*tracemgmtapi.JaegerConfigInfo)(nil)),
			HandlerFunc:  tracingStatusHandler},
		route.Route{
			Name:         "TraceUpdate",
			Method:       "POST",
			Pattern:      "/tracemgmt/update",
			Version:      1,
			RequestType:  utils.GetTypeString((*tracemgmtapi.SetupTracingReq)(nil)),
			ResponseType: utils.GetTypeString((*tracemgmtapi.JaegerConfigInfo)(nil)),
			HandlerFunc:  tracingUpdateHandler},
		route.Route{
			Name:        "TraceDisable",
			Method:      "DELETE",
			Pattern:     "/tracemgmt",
			Version:     1,
			HandlerFunc: tracingDisableHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with Glusterd transaction framework
func (p *Plugin) RegisterStepFuncs() {
	oldtransaction.RegisterStepFunc(txnTracingValidateConfig, "trace-mgmt.ValidateTraceConfig")
	oldtransaction.RegisterStepFunc(txnTracingStoreConfig, "trace-mgmt.StoreTraceConfig")
	oldtransaction.RegisterStepFunc(txnTracingUndoStoreConfig, "trace-mgmt.RestoreTraceConfig")
	oldtransaction.RegisterStepFunc(txnTracingDeleteStoreConfig, "trace-mgmt.UndoStoreTraceConfig")
	oldtransaction.RegisterStepFunc(txnTracingApplyNewConfig, "trace-mgmt.NotifyTraceConfigChange")
	oldtransaction.RegisterStepFunc(txnTracingDisable, "trace-mgmt.NotifyTraceDisable")
}
