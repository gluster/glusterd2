// Package volumecreate implements the volume create command for GlusterD
package volumecreate

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

// Command is a holding struct used to implement the GlusterD Command interface
// for the volume create command
type Command struct {
}

func validateVolumeCreateRequest(msg *volume.VolCreateRequest, r *http.Request) *client.GenericJSONResponse {
	e := utils.GetJSONFromRequest(r, msg)
	if e != nil {
		log.Error("Invalid JSON Request")
		return client.FormResponse(-1, 422, errors.ErrJSONParsingFailed.Error(), "")
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		return client.FormResponse(-1, http.StatusBadRequest, errors.ErrEmptyVolName.Error(), "")
	}
	if len(msg.Bricks) <= 0 {
		log.Error("Brick list is empty")
		return client.FormResponse(-1, http.StatusBadRequest, errors.ErrEmptyBrickList.Error(), "")
	}
	return nil

}

func createVolume(msg *volume.VolCreateRequest) *volume.Volinfo {
	vol := volume.NewVolumeEntry(msg)
	return vol
}

func (c *Command) volumeCreate(w http.ResponseWriter, r *http.Request) {

	msg := new(volume.VolCreateRequest)

	rsp := validateVolumeCreateRequest(msg, r)
	if rsp != nil {
		client.SendResponse(w, http.StatusBadRequest, rsp)
		return
	}
	if context.Store.VolumeExists(msg.Name) {
		log.WithField("Volume", msg.Name).Error("Volume already exists")
		rsp = client.FormResponse(-1, http.StatusBadRequest, errors.ErrVolExists.Error(), "")
		client.SendResponse(w, http.StatusBadRequest, rsp)
		return
	}
	vol := createVolume(msg)
	if vol == nil {
		rsp = client.FormResponse(-1, http.StatusBadRequest, errors.ErrVolCreateFail.Error(), "")
		client.SendResponse(w, http.StatusBadRequest, rsp)
		return
	}

	//TODO : Error handling for volgen
	// Creating client  and server volfile
	e := volgen.GenerateVolfile(vol)
	if e != nil {
		rsp = client.FormResponse(-1, http.StatusInternalServerError, e.Error(), "")
		client.SendResponse(w, http.StatusInternalServerError, rsp)
		return
	}

	e = volume.AddOrUpdateVolume(vol)
	if e != nil {
		rsp = client.FormResponse(-1, http.StatusInternalServerError, e.Error(), "")
		client.SendResponse(w, http.StatusInternalServerError, rsp)
		return
	}

	log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	rsp = client.FormResponse(0, 0, "", "")
	client.SendResponse(w, http.StatusCreated, rsp)
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
