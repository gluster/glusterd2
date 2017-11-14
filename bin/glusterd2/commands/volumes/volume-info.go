package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/bin/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/bin/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

func volumeInfoHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	vol, e := volume.GetVolume(volname)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusOK, vol)
	}
}
