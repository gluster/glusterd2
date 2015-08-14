// Package volumestart implements the volume start command for GlusterD
package volumestart

import (
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume start command
type Command struct {
}

func (c *Command) volumeStart(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume start API")

	vol, e := context.Store.GetVolume(volname)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	if vol.Status == volume.VolStarted {
		http.Error(w, "Volume is already started", http.StatusBadRequest)
		return
	}
	vol.Status = volume.VolStarted

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

// Routes returns command routes to be set up for the volume start command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeStart
		rest.Route{
			Name:        "VolumeStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/start",
			HandlerFunc: c.volumeStart},
	}
}
