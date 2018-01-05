package volumecommands

import (
	"context"
	"encoding/json"
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
)

func volumeProfileTunablesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp, err := store.Store.Get(context.TODO(), "groupoptions")
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	var groupOptions map[string][]api.Option
	if err := json.Unmarshal(resp.Kvs[0].Value, &groupOptions); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	profilename := mux.Vars(r)["profilename"]
	optionSet, ok := groupOptions[profilename]
	if !ok {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "profile not found", api.ErrCodeDefault)
		return
	}

	var response api.ProfileTunablesResp
	for _, option := range optionSet {
		response.Tunables = append(response.Tunables, option.OptionName)
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, response)
}
