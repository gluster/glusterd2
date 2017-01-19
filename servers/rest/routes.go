package rest

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// Route models a route to be set on the GlusterD Rest server
// This route style comes from the tutorial on
// http://thenewstack.io/make-a-restful-json-api-go/
type Route struct {
	Name        string
	Method      string
	Pattern     string
	Version     int
	HandlerFunc http.HandlerFunc
}

// Routes is a table of many Route's
type Routes []Route

// SetRoutes adds the given routes to the GlusterD Rest server
func (r *GDRest) SetRoutes(routes Routes) {
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
