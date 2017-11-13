// Package versioncommands implements the version ReST end point
package versioncommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/bin/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/version"
)

// VersionResponse represents the structure of the response object for
// /version end point. TODO: Move this 'pkg/api'
type VersionResponse struct {
	GlusterdVersion string
	APIVersion      int
}

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	v := VersionResponse{
		GlusterdVersion: version.GlusterdVersion,
		APIVersion:      version.APIVersion,
	}
	restutils.SendHTTPResponse(w, http.StatusOK, v)
}
