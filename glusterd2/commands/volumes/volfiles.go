package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	volgen "github.com/gluster/glusterd2/glusterd2/volgen2"
	"github.com/gluster/glusterd2/pkg/api"
)

func volfilesGenerateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := volgen.Generate()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "unable to generate volfiles", api.ErrCodeDefault)
		return
	}
	volfiles, err := volgen.GetVolfiles()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "unable to get list of volfiles", api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volfiles)
}

func volfilesListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	volfiles, err := volgen.GetVolfiles()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "unable to get list of volfiles", api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volfiles)
}
