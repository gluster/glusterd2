// Package versioncommands implements the version ReST end point
package versioncommands

import (
	"net/http"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/utils"
)

// Response represents the structure of the response object for /version
// end point
type Response struct {
	OpVersion  int
	APIVersion int
}

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	var v Response
	v.OpVersion = context.GlusterdVersion
	v.APIVersion = context.APIVersion
	utils.SendHTTPResponse(w, http.StatusOK, v)
}
