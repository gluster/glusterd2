// Package volumelist implements the volume list command for GlusterD
package volumelist

import (
	"encoding/json"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"

	log "github.com/Sirupsen/logrus"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume info command
type Command struct {
}

func (c *Command) volumeList(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	volumes, e := context.Store.GetVolumes()
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	// Write nsg
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if e = json.NewEncoder(w).Encode(volumes); e != nil {
		panic(e)
	}
}

// Routes returns command routes to be set up for the volume info command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeList
		rest.Route{
			Name:        "VolumeList",
			Method:      "GET",
			Pattern:     "/volumes/",
			HandlerFunc: c.volumeList},
	}
}
