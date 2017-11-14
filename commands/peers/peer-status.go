package peercommands

// TODO: Reimplement these endpoints later

import (
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
)

func peerEtcdStatusHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPError(w, http.StatusNotFound, "", api.ErrCodeDefault)
}

func peerEtcdHealthHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPError(w, http.StatusNotFound, "", api.ErrCodeDefault)
}
