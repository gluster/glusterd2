package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func volumeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume delete API")

	if volume.Exists(volname) {
		client.SendResponse(w, -1, http.StatusBadRequest, errors.ErrVolNotFound.Error(), http.StatusBadRequest, "")
		return
	}

	e := volume.DeleteVolume(volname)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": volname,
		}).Error("Failed to delete the volume")
		client.SendResponse(w, -1, http.StatusInternalServerError, e.Error(), http.StatusInternalServerError, "")
		return
	}
	client.SendResponse(w, 0, 0, "", http.StatusOK, "")
}
