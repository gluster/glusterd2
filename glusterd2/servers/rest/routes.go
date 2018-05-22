package rest

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/commands"
	"github.com/gluster/glusterd2/glusterd2/plugin"
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

// AllRoutes is global list of all API endpoints
var AllRoutes route.Routes

// setRoutes adds the given routes to the GlusterD Rest server
func (r *GDRest) setRoutes(routes route.Routes) {
	var urlPattern string
	for _, route := range routes {
		// Set routes in mux.Routes
		if route.Version == 0 {
			urlPattern = route.Pattern
		} else {
			urlPattern = fmt.Sprintf("/v%d%s", route.Version, route.Pattern)
		}
		log.WithFields(log.Fields{
			"name":   route.Name,
			"path":   urlPattern,
			"method": route.Method,
		}).Debug("Registering new mux route")
		r.Routes.
			Methods(route.Method).
			Path(urlPattern).
			Name(route.Name).
			Handler(route.HandlerFunc)

		// Set our global copy of all routes
		AllRoutes = append(AllRoutes, route)
	}
}

func (r *GDRest) registerRoutes() {
	for _, c := range commands.Commands {
		r.setRoutes(c.Routes())
		//XXX: This doesn't feel like the right place to be register step
		//functions, but until we have a better place it can stay here.
		c.RegisterStepFuncs()
	}

	// Load routes and Step functions from Plugins
	for _, p := range plugin.PluginsList {
		restRoutes := p.RestRoutes()
		if restRoutes != nil {
			r.setRoutes(restRoutes)
			log.WithField("plugin", p.Name()).Debug("loaded REST routes from plugin")
		}
		p.RegisterStepFuncs()
	}

	// Expose /statedump and /endpoints handlers
	var moreRoutes route.Routes

	if ok := config.GetBool("statedump"); ok {
		moreRoutes = append(moreRoutes, route.Route{
			Name:        "Statedump",
			Method:      "GET",
			Pattern:     "/statedump",
			HandlerFunc: expvar.Handler().(http.HandlerFunc)})
	}

	moreRoutes = append(moreRoutes, route.Route{
		Name:         "List Endpoints",
		Method:       "GET",
		Pattern:      "/endpoints",
		ResponseType: utils.GetTypeString((*api.ListEndpointsResp)(nil)),
		HandlerFunc:  r.listEndpointsHandler()})

	moreRoutes = append(moreRoutes, route.Route{
		Name:        "Glusterd2 service status",
		Method:      "GET",
		Pattern:     "/ping",
		HandlerFunc: r.Ping()})
	r.setRoutes(moreRoutes)
}
