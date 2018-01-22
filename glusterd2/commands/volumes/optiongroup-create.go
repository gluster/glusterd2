package volumecommands

import (
	"context"
	"encoding/json"
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

func optionGroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.OptionGroupReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

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

	if _, ok := groupOptions[req.Name]; ok {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "profile already exists", api.ErrCodeDefault)
		return
	}

	var optionSet []api.VolumeOption
	for _, option := range req.Options {
		optionSet = append(optionSet, option)
	}
	groupOptions[req.Name] = optionSet

	groupOptionsJSON, err := json.Marshal(groupOptions)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	if _, err := store.Store.Put(context.TODO(), "groupoptions", string(groupOptionsJSON)); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, req)
}
