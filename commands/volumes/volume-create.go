package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func validateVolumeCreateJSONRequest(msg *volume.VolCreateRequest, r *http.Request) (int, error) {
	e := utils.GetJSONFromRequest(r, msg)
	if e != nil {
		log.WithField("error", e).Error("Failed to parse the JSON Request")
		return 422, errors.ErrJSONParsingFailed
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		return http.StatusBadRequest, errors.ErrEmptyVolName
	}
	if len(msg.Bricks) <= 0 {
		log.WithField("volume", msg.Name).Error("Brick list is empty")
		return http.StatusBadRequest, errors.ErrEmptyBrickList
	}
	return 0, nil

}

func createVolume(msg *volume.VolCreateRequest) (*volume.Volinfo, error) {
	vol, err := volume.NewVolumeEntry(msg)
	if err != nil {
		return nil, err
	}
	vol.Bricks, err = volume.NewBrickEntries(msg.Bricks)
	if err != nil {
		return nil, err
	}
	return vol, nil
}

func validateVolumeCreate(msg *volume.VolCreateRequest, v *volume.Volinfo) (int, error) {
	if volume.Exists(msg.Name) {
		log.WithField("volume", msg.Name).Error("Volume already exists")
		return http.StatusBadRequest, errors.ErrVolExists
	}
	httpStatusCode, err := volume.ValidateBrickEntries(v.Bricks, v.ID, msg.Force)
	if err != nil {
		return httpStatusCode, err
	}
	return 0, nil
}

func commitVolumeCreate(vol *volume.Volinfo) (int, error) {
	// Creating client and server volfile
	e := volgen.GenerateVolfile(vol)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to generate volfile")
		return http.StatusInternalServerError, e
	}

	e = volume.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to create volume")
		return http.StatusInternalServerError, e
	}
	log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	return 0, nil
}

func rollBackVolumeCreate(vol *volume.Volinfo) error {
	volume.RemoveBrickPaths(vol.Bricks)
	return nil
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	log.Debug("In volume create")
	msg := new(volume.VolCreateRequest)

	httpStatusCode, e := validateVolumeCreateJSONRequest(msg, r)
	if e != nil {
		rest.SendHTTPError(w, httpStatusCode, e.Error())
		return
	}
	vol, e := createVolume(msg)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	httpStatusCode, e = validateVolumeCreate(msg, vol)
	if e != nil {
		rest.SendHTTPError(w, httpStatusCode, e.Error())
		return
	}
	httpStatusCode, e = commitVolumeCreate(vol)
	if e != nil {
		rollBackVolumeCreate(vol)
		rest.SendHTTPError(w, httpStatusCode, e.Error())
		return
	}
	rest.SendHTTPResponse(w, http.StatusCreated, vol)
}
