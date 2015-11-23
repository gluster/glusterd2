package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/errors"
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
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), http.StatusBadRequest, "")
		return
	}
	if vol.Status == volume.VolStopped {
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolAlreadyStopped.Error(), http.StatusBadRequest, "")
		return
	}
	vol.Status = volume.VolStopped

	e = volume.AddOrUpdateVolume(vol)
	if e != nil {
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}
	log.WithField("volume", vol.Name).Debug("Volume updated into the store")
	client.SendResponse(w, 0, 0, "", http.StatusOK, vol)
}
