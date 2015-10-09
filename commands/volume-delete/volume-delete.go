// Package volumedelete implements the volume delete command for GlusterD
package volumedelete

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
// for the volume delete command
type Command struct {
}

func (c *Command) volumeDelete(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume delete API")

	if context.Store.VolumeExists(volname) {
		rsp := client.FormResponse(-1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), "")
		client.SendResponse(w, http.StatusBadRequest, rsp)
		return
	}

	e := volume.DeleteVolume(volname)
	if e != nil {
		rsp := client.FormResponse(-1, http.StatusInternalServerError, e.Error(), "")
		client.SendResponse(w, http.StatusInternalServerError, rsp)
		return
	}
	rsp := client.FormResponse(0, 0, "", "")
	client.SendResponse(w, http.StatusOK, rsp)
}

// Routes returns command routes to be set up for the volume delete command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeDelete
		rest.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: c.volumeDelete},
	}
}
