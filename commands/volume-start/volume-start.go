// Package volumestart implements the volume start command for GlusterD
package volumestart

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume start command
type Command struct {
}

func (c *Command) volumeStart(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	/* TODO : As of now we consider the request as volname, later on we need
	* to consider the volume id as well */
	volname := p["volname"]

	log.Info("In Volume start API")

	vol, e := context.Store.GetVolume(volname)
	if e != nil {
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), http.StatusBadRequest, "")
		return
	}
	if vol.Status == volume.VolStarted {
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolAlreadyStarted.Error(), http.StatusBadRequest, "")
		return
	}
	vol.Status = volume.VolStarted

	e = context.Store.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't update volume into the store")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	client.SendResponse(w, 0, 0, "", http.StatusOK, "")
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
