package rest

import (
	"net/http"
)

//
// This route style comes from the tutorial on
// http://thenewstack.io/make-a-restful-json-api-go/
//
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route
