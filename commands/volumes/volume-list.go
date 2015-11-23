package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/client"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	volumes, e := volume.GetVolumes()

	if e != nil {
		client.SendResponse(w, -1, http.StatusNotFound, e.Error(), http.StatusNotFound, "")
	} else {
		client.SendResponse(w, 0, 0, "", http.StatusOK, volumes)
	}
}
