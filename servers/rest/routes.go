package rest

import (
	"fmt"

	"github.com/gluster/glusterd2/servers/rest/route"

	log "github.com/Sirupsen/logrus"
)

// SetRoutes adds the given routes to the GlusterD Rest server
func (r *GDRest) SetRoutes(routes route.Routes) {
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
