package rest

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/commands"
	"github.com/gluster/glusterd2/glusterd2/plugin"
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"

	log "github.com/sirupsen/logrus"
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
}
