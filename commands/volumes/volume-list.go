package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	volumes, e := volume.GetVolumes()

	if e != nil {
		rest.SendHTTPError(w, http.StatusNotFound, e.Error())
	} else {
		rest.SendHTTPResponse(w, http.StatusOK, volumes)
	}
}
