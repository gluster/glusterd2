package rest

import (
	"fmt"

	"github.com/gluster/glusterd2/bin/glusterd2/commands"
	"github.com/gluster/glusterd2/plugins"
	"github.com/gluster/glusterd2/servers/rest/route"

	log "github.com/sirupsen/logrus"
)

// setRoutes adds the given routes to the GlusterD Rest server
func (r *GDRest) setRoutes(routes route.Routes) {
	for _, route := range routes {
		var urlPattern string
		if route.Name == "GetVersion" {
			urlPattern = route.Pattern
		} else {
			urlPattern = fmt.Sprintf("/v%d%s", route.Version, route.Pattern)
		}
		log.WithFields(log.Fields{
			"name":   route.Name,
			"path":   urlPattern,
			"method": route.Method,
		}).Debug("Registering new route")

		r.Routes.
			Methods(route.Method).
			Path(urlPattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
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
	for _, p := range plugins.PluginsList {
		restRoutes := p.RestRoutes()
		if restRoutes != nil {
			r.setRoutes(restRoutes)
			log.WithField("plugin", p.Name()).Debug("loaded REST routes from plugin")
		}
		p.RegisterStepFuncs()
	}
}
