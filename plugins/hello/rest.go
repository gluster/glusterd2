package hello

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/servers/rest/utils"
)

func helloGetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Hello Get")
}

func helloPostHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Hello Post")
}
