package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func volumeInfoHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Debug("In Volume info API")

	// Simple read operations, which just read information from the store, need
	// not use the transaction framework
	vol, e := volume.GetVolume(volname)
	if e != nil {
		rest.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
	} else {

		rest.SendHTTPResponse(w, http.StatusOK, vol)
	}
}
