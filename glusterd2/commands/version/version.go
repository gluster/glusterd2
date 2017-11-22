// Package versioncommands implements the version ReST end point
package versioncommands

import (
	"net/http"

	"github.com/gluster/glusterd2/version"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
)

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := api.VersionResp{
		GlusterdVersion: version.GlusterdVersion,
		APIVersion:      version.APIVersion,
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}
