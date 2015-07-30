package volumecreate

import (
	"fmt"
	"net/http"

	"github.com/kshlm/glusterd2/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeCreateCommand struct {
}

func (c *VolumeCreateCommand) VolumeCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Volume Create API")
	log.Info("In Volume create API")
}

func (c *VolumeCreateCommand) SetRoutes(router *mux.Router) error {
	routes := rest.Routes{
		// VolumeCreate
		rest.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes",
			HandlerFunc: c.VolumeCreate},
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
