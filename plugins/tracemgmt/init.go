package tracemgmt

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
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
	}
}

// RegisterStepFuncs registers transaction step functions with Glusterd transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnTracingValidateConfig, "trace-mgmt.ValidateTraceConfig")
	transaction.RegisterStepFunc(txnTracingStoreConfig, "trace-mgmt.StoreTraceConfig")
	transaction.RegisterStepFunc(txnTracingDeleteStoreConfig, "trace-mgmt.UndoStoreTraceConfig")
	transaction.RegisterStepFunc(txnTracingApplyNewConfig, "trace-mgmt.NotifyTraceConfigChange")
}
