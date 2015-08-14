// Package volumeinfo implements the volume info command for GlusterD
package volumeinfo

import (
	"net/http"

	"github.com/kshlm/glusterd2/cli"
	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/errors"
	"github.com/kshlm/glusterd2/rest"

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

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vol, e := context.Store.GetVolume(volname)
	if e != nil {
		cli.SendResponse(w, -1, http.StatusNotFound, errors.ErrVolNotFound.Error(), http.StatusNotFound, "")
	} else {
		cli.SendResponse(w, 0, http.StatusOK, "", http.StatusNotFound, vol)
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
