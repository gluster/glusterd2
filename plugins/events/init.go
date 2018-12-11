package events

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "events"
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "EventsWebhookAdd",
			Method:      "POST",
			Pattern:     "/events/webhook",
			Version:     1,
			RequestType: utils.GetTypeString((*eventsapi.Webhook)(nil)),
			HandlerFunc: webhookAddHandler},
		route.Route{
			Name:        "EventsWebhookTest",
			Method:      "POST",
			Pattern:     "/events/webhook/test",
			Version:     1,
			RequestType: utils.GetTypeString((*eventsapi.Webhook)(nil)),
			HandlerFunc: webhookTestHandler},
		route.Route{
			Name:        "EventsWebhookDelete",
			Method:      "DELETE",
			Pattern:     "/events/webhook",
			Version:     1,
			RequestType: utils.GetTypeString((*eventsapi.WebhookDel)(nil)),
			HandlerFunc: webhookDeleteHandler},
		route.Route{
			Name:         "EventsWebhookList",
			Method:       "GET",
			Pattern:      "/events/webhook",
			Version:      1,
			ResponseType: utils.GetTypeString((*eventsapi.WebhookList)(nil)),
			HandlerFunc:  webhookListHandler},
		route.Route{
			Name:    "EventsList",
			Method:  "GET",
			Pattern: "/events",
			Version: 1,
			// FIXME: This type is not in 'eventsapi'
			ResponseType: utils.GetTypeString((*api.Event)(nil)),
			HandlerFunc:  eventsListHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	registerWebhookTestStepFuncs()
	return
}
