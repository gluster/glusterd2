package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	volumes, e := volume.GetVolumes()

	if e != nil {
		utils.SendHTTPError(w, http.StatusNotFound, e.Error())
	} else {
		utils.SendHTTPResponse(w, http.StatusOK, volumes)
	}
}
