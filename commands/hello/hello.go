// Package hello implements a dummy Hello Command
package hello

import (
	"fmt"
	"github.com/kshlm/glusterd2/rest"
	"net/http"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

func (h *Command) hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "HelloWorld from GlusterFS Application")
}

// Routes returns command routes. Required for the Command interface.
func (h *Command) Routes() rest.Routes {
	return rest.Routes{
		// HelloWorld
		rest.Route{
			Name:        "Hello",
			Method:      "GET",
			Pattern:     "/hello",
			HandlerFunc: h.hello},
	}
}
