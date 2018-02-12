package events

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/sunrpc"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "events"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "WebhookAdd",
			Method:      "POST",
			Pattern:     "/events/webhook",
			Version:     1,
			HandlerFunc: webhookAddHandler},
		route.Route{
			Name:        "WebhookDelete",
			Method:      "DELETE",
			Pattern:     "/events/webhook",
			Version:     1,
			HandlerFunc: webhookDeleteHandler},
		route.Route{
			Name:        "WebhookList",
			Method:      "GET",
			Pattern:     "/events/webhook",
			Version:     1,
			HandlerFunc: webhookListHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	return
}
