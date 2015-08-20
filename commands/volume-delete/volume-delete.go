// Package volumedelete implements the volume delete command for GlusterD
package volumedelete

import (
	"net/http"

	"github.com/kshlm/glusterd2/client"
	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/errors"
	"github.com/kshlm/glusterd2/rest"

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
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), http.StatusBadRequest, "")
		return
	}

	e := context.Store.DeleteVolume(volname)
	if e != nil {
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
			HandlerFunc: c.volumeDelete},
	}
}
