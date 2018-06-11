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

func optionGroupDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groupName := mux.Vars(r)["groupname"]

	resp, err := store.Get(context.TODO(), "groupoptions")
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	var groupOptions map[string]*api.OptionGroup
	if err := json.Unmarshal(resp.Kvs[0].Value, &groupOptions); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	_, ok := groupOptions[groupName]
	if !ok {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid group name specified")
		return
	}

	if _, ok := defaultGroupOptions[groupName]; ok {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "cannot delete builtin groups")
		return
	}

	delete(groupOptions, groupName)

	groupOptionsJSON, err := json.Marshal(groupOptions)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if _, err := store.Put(context.TODO(), "groupoptions", string(groupOptionsJSON)); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}
