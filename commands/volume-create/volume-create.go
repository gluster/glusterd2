package volumecreate

import (
	"fmt"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/utils"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeCreateCommand struct {
}

func (c *VolumeCreateCommand) VolumeCreate(w http.ResponseWriter, r *http.Request) {

	var msg volume.VolumeCreateRequest

	e := utils.GetJsonFromRequest(r, &msg)
	if e != nil {
		http.Error(w, "request unable to be parsed", 422)
		return
	}

	if len(msg.Name) <= 0 {
		log.Error("Volume name is empty")
		http.Error(w, "Volume name is empty", http.StatusBadRequest)

		return
	}
	if len(msg.Bricks) <= 0 {
		log.Error("Brick list is empty")
		http.Error(w, "Brick list is empty", http.StatusBadRequest)

		return
	}

	if context.Store.VolumeExists(msg.Name) {
		log.WithField("Volume", msg.Name).Error("Volume already exists")
		http.Error(w, "Volume already exists", http.StatusBadRequest)

		return
	}

	vol := volume.New(msg.Name, msg.Transport, msg.ReplicaCount,
		msg.StripeCount, msg.DisperseCount,
		msg.RedundancyCount, msg.Bricks)

	e = context.Store.AddVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't add volume to store")
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	} else {
		log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Volume created successfully")
}

func (c *VolumeCreateCommand) SetRoutes(router *mux.Router) error {
	routes := rest.Routes{
		// VolumeCreate
		rest.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes/",
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
