package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func validateVolumeCreateRequest(msg *volume.VolCreateRequest, r *http.Request, w http.ResponseWriter) error {
	e := utils.GetJSONFromRequest(r, msg)
	if e != nil {
		log.WithField("error", e).Error("Failed to parse the JSON Request")
		client.SendResponse(w, -1, 422, errors.ErrJSONParsingFailed.Error(), 422, "")
		return errors.ErrJSONParsingFailed
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrEmptyVolName.Error(), http.StatusBadRequest, "")
		return errors.ErrEmptyVolName
	}
	if len(msg.Bricks) <= 0 {
		log.WithField("volume", msg.Name).Error("Brick list is empty")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrEmptyBrickList.Error(), http.StatusBadRequest, "")
		return errors.ErrEmptyBrickList
	}
	return nil

}

func createVolume(msg *volume.VolCreateRequest) *volume.Volinfo {
	vol := volume.NewVolumeEntry(msg)
	return vol
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	msg := new(volume.VolCreateRequest)

	e := validateVolumeCreateRequest(msg, r, w)
	if e != nil {
		// Response has been already sent, just return
		return
	}
	if volume.Exists(msg.Name) {
		log.WithField("volume", msg.Name).Error("Volume already exists")
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolExists.Error(), http.StatusBadRequest, "")
		return
	}
	vol := createVolume(msg)
	if vol == nil {
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolCreateFail.Error(), http.StatusBadRequest, "")
		return
	}

	// Creating client  and server volfile
	e = volgen.GenerateVolfile(vol)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to generate volfile")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}

	e = volume.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to create volume")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}

	log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	client.SendResponse(w, 0, 0, "", http.StatusCreated, vol)
}
