package hello

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kshlm/glusterd2/rest"
	"net/http"
)

type HelloCommand struct {
}

func (h *HelloCommand) Hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "HelloWorld from GlusterFS Application")
}

func (h *HelloCommand) SetRoutes(router *mux.Router) error {
	routes := rest.Routes{
		// HelloWorld
		rest.Route{
			Name:        "Hello",
			Method:      "GET",
			Pattern:     "/hello",
			HandlerFunc: h.Hello},
	}
	// Register all routes
	for _, route := range routes {
		// Add routes from the table
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return nil

}
