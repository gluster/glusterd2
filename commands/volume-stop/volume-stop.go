package volumestop

import (
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeStopCommand struct {
}

func (c *VolumeStopCommand) VolumeStop(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume stop API")

	vol, e := context.Store.GetVolume(volname)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	if vol.Status == volume.VolStopped {
		http.Error(w, "Volume is already stopped", http.StatusBadRequest)
		return
	}
	vol.Status = volume.VolStopped

	e = context.Store.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't update volume into the store")
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	} else {
		log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	}

	// Write nsg
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (c *VolumeStopCommand) SetRoutes(router *mux.Router) error {
	routes := rest.Routes{
		// VolumeStop
		rest.Route{
			Name:        "VolumeStop",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/stop",
			HandlerFunc: c.VolumeStop},
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
