package device

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/utils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "device"
}

// RestRoutes returns list of REST API routes to register with Glusterd.
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "DeviceAdd",
			Method:       "POST",
			Pattern:      "/devices/{peerid}",
			Version:      1,
			RequestType:  utils.GetTypeString((*deviceapi.AddDeviceReq)(nil)),
			ResponseType: utils.GetTypeString((*deviceapi.AddDeviceResp)(nil)),
			HandlerFunc:  deviceAddHandler},
		route.Route{
			Name:         "DeviceList",
			Method:       "GET",
			Pattern:      "/devices/{peerid}",
			Version:      1,
			ResponseType: utils.GetTypeString((*deviceapi.ListDeviceResp)(nil)),
			HandlerFunc:  deviceListHandler},
		route.Route{
			Name:        "DeviceEdit",
			Method:      "POST",
			Pattern:     "/devices/{peerid}",
			Version:     1,
			RequestType: utils.GetTypeString((*deviceapi.EditDeviceReq)(nil)),
			HandlerFunc: deviceEditHandler},
		route.Route{
			Name:         "DeviceListAll",
			Method:       "GET",
			Pattern:      "/devices",
			Version:      1,
			ResponseType: utils.GetTypeString((*deviceapi.ListDeviceResp)(nil)),
			HandlerFunc:  listAllDevicesHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	transaction.RegisterStepFunc(txnPrepareDevice, "prepare-device")
}
