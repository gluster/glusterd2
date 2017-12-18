package quota

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"
)

func commonValidator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	volname := mux.Vars(r)["volname"]
	v, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
	} else if v.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted.Error(), api.ErrCodeDefault)
	}
}

func quotaEnableHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "todo: quota enable")
}

func quotaDisableHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "todo: quota disable")
}

func quotaListHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "todo: quota list")
}

func quotaLimitHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Todo: limitusage")
}

func quotaRemoveHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Todo: quota Remove")
}
