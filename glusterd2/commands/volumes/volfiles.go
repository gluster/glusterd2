package volumecommands

import (
	"fmt"
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	volgen "github.com/gluster/glusterd2/glusterd2/volgen2"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"
)

func volfilesGenerateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := volgen.Generate()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "unable to generate volfiles")
		return
	}
	volfiles, err := volgen.GetVolfiles()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "unable to get list of volfiles")
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volfiles)
}

func volfilesListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	volfiles, err := volgen.GetVolfiles()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "unable to get list of volfiles")
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volfiles)
}

func volfileGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	volfileid := mux.Vars(r)["volfileid"]

	volfile, err := volgen.GetVolfile(volfileid)

	if err != nil {
		if err == errors.ErrVolFileNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolFileNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("unable to fetch volfile content %s", volfile))
		}
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write(volfile)
}
