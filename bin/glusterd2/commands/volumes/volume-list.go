package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/volume"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	volumes, e := volume.GetVolumes()
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, e.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusOK, volumes)
	}
}
