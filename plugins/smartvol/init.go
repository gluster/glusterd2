package smartvol

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
	smartvolapi "github.com/gluster/glusterd2/plugins/smartvol/api"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "smartvol"
}

// RestRoutes returns list of REST API routes to register with Glusterd.
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "SmartVolumeCreate",
			Method:       "POST",
			Pattern:      "/smartvol",
			Version:      1,
			RequestType:  utils.GetTypeString((*smartvolapi.VolCreateReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeCreateResp)(nil)),
			HandlerFunc:  smartVolumeCreateHandler,
		},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnPrepareBricks, "vol-create.PrepareBricks")
	transaction.RegisterStepFunc(txnUndoPrepareBricks, "vol-create.UndoPrepareBricks")
}
