// Package volumedelete implements the volume delete command for GlusterD
package volumedelete

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
// for the volume delete command
type Command struct {
}

func (c *Command) volumeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume delete API")

	if volume.VolumeExists(volname) {
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), http.StatusBadRequest, "")
		return
	}

	e := volume.DeleteVolume(volname)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": volname,
		}).Error("Failed to delete the volume")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}
	client.SendResponse(w, 0, 0, "", http.StatusOK, "")
}

// Routes returns command routes to be set up for the volume delete command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeDelete
		rest.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: c.volumeDeleteHandler},
	}
}
