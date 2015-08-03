package volumecreate

import (
	"fmt"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeCreateCommand struct {
}

func (c *VolumeCreateCommand) VolumeCreate(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Volume Create API")
	log.Info("In Volume create API")

	vol := volume.New(volname, "tcp", 2, 3, 4, 5, []string{"brick1", "brick2"})
	e := context.Store.AddVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't add volume to store")
	} else {
		log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	}
}

func (c *VolumeCreateCommand) SetRoutes(router *mux.Router) error {
	routes := rest.Routes{
		// VolumeCreate
		rest.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes/{volname}",
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
