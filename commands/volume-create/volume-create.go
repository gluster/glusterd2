// Package volumecreate implements the volume create command for GlusterD
package volumecreate

import (
	"fmt"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/utils"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume create command
type Command struct {
}

func (c *Command) volumeCreate(w http.ResponseWriter, r *http.Request) {

	var msg volume.VolCreateRequest

	e := utils.GetJSONFromRequest(r, &msg)
	if e != nil {
		http.Error(w, "request unable to be parsed", 422)
		return
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		http.Error(w, "Volume name is empty", http.StatusBadRequest)

		return
	}
	if len(msg.Bricks) <= 0 {
		log.Error("Brick list is empty")
		http.Error(w, "Brick list is empty", http.StatusBadRequest)

		return
	}

	if context.Store.VolumeExists(msg.Name) {
		log.WithField("Volume", msg.Name).Error("Volume already exists")
		http.Error(w, "Volume already exists", http.StatusBadRequest)

		return
	}

	vol := volume.New(msg.Name, msg.Transport, msg.ReplicaCount,
		msg.StripeCount, msg.DisperseCount,
		msg.RedundancyCount, msg.Bricks)

	e = context.Store.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't add volume to store")
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	log.WithField("volume", vol.Name).Debug("NewVolume added to store")

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Volume created successfully")
}

// Routes returns command routes to be set up for the volume create command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeCreate
		rest.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes/",
			HandlerFunc: c.volumeCreate},
	}
}
