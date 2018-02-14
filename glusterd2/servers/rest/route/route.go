// Package route implements a Route type to be used by GD2 rest API to define their routes
package route

import (
	"net/http"
)

// Route models a route to be set on the GlusterD Rest server
// This route style comes from the tutorial on
// http://thenewstack.io/make-a-restful-json-api-go/
type Route struct {
	Name         string
	Description  string
	Method       string
	Pattern      string
	Version      int
	RequestType  string
	ResponseType string // Success
	HandlerFunc  http.HandlerFunc
}

// Routes is a table of many Route's
type Routes []Route
