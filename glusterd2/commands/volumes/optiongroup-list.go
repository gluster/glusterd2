package volumecommands

import (
	"context"
	"encoding/json"
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
)

func optionGroupListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp, err := store.Store.Get(context.TODO(), "groupoptions")
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	var groupOptions map[string][]api.VolumeOption
	if err := json.Unmarshal(resp.Kvs[0].Value, &groupOptions); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	var response []api.OptionGroup
	for name, options := range groupOptions {
		response = append(response, api.OptionGroup{Name: name, Options: options})
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, response)
}
