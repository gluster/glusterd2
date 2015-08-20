// Package volumecreate implements the volume create command for GlusterD
package volumecreate

import (
	"net/http"

	"github.com/kshlm/glusterd2/client"
	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/errors"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/utils"
	"github.com/kshlm/glusterd2/volgen"
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
		client.SendResponse(w, -1, 422, errors.ErrJSONParsingFailed.Error(), 422, "")
		return
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrEmptyVolName.Error(), http.StatusBadRequest, "")
		return
	}
	if len(msg.Bricks) <= 0 {
		log.Error("Brick list is empty")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrEmptyBrickList.Error(), http.StatusBadRequest, "")
		return
	}

	if context.Store.VolumeExists(msg.Name) {
		log.WithField("Volume", msg.Name).Error("Volume already exists")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolExists.Error(), http.StatusBadRequest, "")
		return
	}

	vol := volume.New(msg.Name, msg.Transport, msg.ReplicaCount,
		msg.StripeCount, msg.DisperseCount,
		msg.RedundancyCount, msg.Bricks)

	e = context.Store.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't add volume to store")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}

	// Creating client  and server volfile
	volgen.GenerateVolfile(vol)

	log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	client.SendResponse(w, 0, 0, "", http.StatusCreated, "")
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
