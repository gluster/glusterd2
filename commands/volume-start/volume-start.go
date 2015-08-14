package volumestart

import (
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeStartCommand struct {
}

func (c *VolumeStartCommand) VolumeStart(w http.ResponseWriter, r *http.Request) {
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
	} else {
		log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	}

	// Write nsg
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (c *VolumeStartCommand) Routes() rest.Routes {
	return rest.Routes{
		// VolumeStart
		rest.Route{
			Name:        "VolumeStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/start",
			HandlerFunc: c.VolumeStart},
	}
}
