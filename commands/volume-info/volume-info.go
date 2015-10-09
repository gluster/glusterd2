// Package volumeinfo implements the volume info command for GlusterD
package volumeinfo

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
// for the volume info command
type Command struct {
}

func (c *Command) volumeInfo(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Debug("In Volume info API")

	vol, e := volume.GetVolume(volname)
	if e != nil {
		rsp := client.FormResponse(-1, http.StatusNotFound, errors.ErrVolNotFound.Error(), "")
		client.SendResponse(w, http.StatusNotFound, rsp)
	} else {
		rsp := client.FormResponse(0, 0, "", vol)
		client.SendResponse(w, http.StatusOK, rsp)
	}
}

// Routes returns command routes to be set up for the volume info command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeInfo
		rest.Route{
			Name:        "VolumeInfo",
			Method:      "GET",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: c.volumeInfo},
	}
}
