// Package volumestop implements the volume stop command for GlusterD
package volumestop

import (
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume stop command
type Command struct {
}

func (c *Command) volumeStop(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	/* TODO : As of now we consider the request as volname, later on we need
	* to consider the volume id as well */
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
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")

	// Write nsg
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// Routes returns command routes to be set up for the volume stop command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeStop
		rest.Route{
			Name:        "VolumeStop",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/stop",
			HandlerFunc: c.volumeStop},
	}
}
