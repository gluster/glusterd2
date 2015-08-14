package hello

import (
	"fmt"
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

func (h *HelloCommand) Routes() rest.Routes {
	return rest.Routes{
		// HelloWorld
		rest.Route{
			Name:        "Hello",
			Method:      "GET",
			Pattern:     "/hello",
			HandlerFunc: h.Hello},
	}
}
