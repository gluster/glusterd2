package rest

import (
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
	HandlerFunc http.HandlerFunc
}

// Routes is a table of many Route's
type Routes []Route

// SetRoutes adds the given routes to the GlusterD Rest server
func (r *GDRest) SetRoutes(routes Routes) {
	for _, route := range routes {
		log.WithFields(log.Fields{
			"name":   route.Name,
			"path":   route.Pattern,
			"method": route.Method,
		}).Debug("Registering new route")

		r.Routes.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
}
