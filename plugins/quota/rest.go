package quota

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
)

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
