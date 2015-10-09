// Package volumestart implements the volume start command for GlusterD
package volumestart

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
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
	volname := p["volname"]

	log.Info("In Volume start API")

	vol, e := volume.GetVolume(volname)
	if e != nil {
		rsp := client.FormResponse(-1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), "")
		client.SendResponse(w, http.StatusBadRequest, rsp)
		return
	}
	if vol.Status == volume.VolStarted {
		rsp := client.FormResponse(-1, http.StatusBadRequest, errors.ErrVolAlreadyStarted.Error(), "")
		client.SendResponse(w, http.StatusBadRequest, rsp)
		return
	}
	vol.Status = volume.VolStarted

	e = volume.AddOrUpdateVolume(vol)
	if e != nil {
		rsp := client.FormResponse(-1, http.StatusInternalServerError, e.Error(), "")
		client.SendResponse(w, http.StatusInternalServerError, rsp)
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	rsp := client.FormResponse(0, 0, "", "")
	client.SendResponse(w, http.StatusOK, rsp)
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
