package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func volumeStopHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume stop API")

	vol, e := volume.GetVolume(volname)
	if e != nil {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolNotFound.Error())
		return
	}
	if vol.Status == volume.VolStopped {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolAlreadyStopped.Error())
		return
	}
	vol.Status = volume.VolStopped

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	rest.SendHTTPResponse(w, http.StatusOK, vol)
}
