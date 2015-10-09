// Package volumelist implements the volume list command for GlusterD
package volumelist

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume info command
type Command struct {
}

func (c *Command) volumeList(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	volumes, e := volume.GetVolumes()

	if e != nil {
		rsp := client.FormResponse(-1, http.StatusNotFound, e.Error(), "")
		client.SendResponse(w, http.StatusNotFound, rsp)
	} else {
		rsp := client.FormResponse(0, 0, "", volumes)
		client.SendResponse(w, http.StatusOK, rsp)
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
