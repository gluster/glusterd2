package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	// Simple read operations, which just read information from the store, need
	// not use the transaction framework
	volumes, e := volume.GetVolumesList()

	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, e.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusOK, volumes)
	}
}
