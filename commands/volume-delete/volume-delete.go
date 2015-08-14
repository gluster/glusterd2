// Package volumedelete implements the volume delete command for GlusterD
package volumedelete

import (
	"fmt"
	"net/http"

	"github.com/kshlm/glusterd2/context"
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

	log.Info("In Volume info API")

	e := context.Store.DeleteVolume(volname)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	// Write nsg
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Volume deleted successfully")
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
