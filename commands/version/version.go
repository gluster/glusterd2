// Package versioncommands implements the version ReST end point
package versioncommands

import (
	"net/http"

	"github.com/gluster/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
)

// Response represents the structure of the response object for /version
// end point
type Response struct {
	GlusterdVersion string
	APIVersion      int
}

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	var v Response
	v.GlusterdVersion = gdctx.GlusterdVersion
	v.APIVersion = gdctx.APIVersion
	restutils.SendHTTPResponse(w, http.StatusOK, v)
}
