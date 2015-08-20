// Package volumeaddbrick implements the volume add-brick command for GlusterD
package volumeaddbrick

import (
	"net/http"

	"github.com/kshlm/glusterd2/client"
	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/errors"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/utils"
	"github.com/kshlm/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume add-brick command
type Command struct {
}

func (c *Command) volumeAddBrick(w http.ResponseWriter, r *http.Request) {

	var msg volume.VolAddBrickRequest
	var vol *volume.Volinfo

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

	vol, e = context.Store.GetVolume(msg.Name)
	if e != nil {
		log.WithField("Volume", msg.Name).Error("Volume doesn't exist")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), http.StatusBadRequest, "")
		return
	}
	//TODO : Add logic to match the existing bricks and then append

	e = context.Store.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithField("error", e).Error("Couldn't update volume to store")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}

	log.WithField("volume", vol.Name).Debug("Volume updated into store")
	client.SendResponse(w, 0, 0, "", http.StatusOK, "")
}

// Routes returns command routes to be set up for the volume add-brick command.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		// VolumeAddBrick
		rest.Route{
			Name:        "VolumeAddBrick",
			Method:      "POST",
			Pattern:     "/volumes/add-brick",
			HandlerFunc: c.volumeAddBrick},
	}
}
