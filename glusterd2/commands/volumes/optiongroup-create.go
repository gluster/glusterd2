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

func validateOptionSet(req api.OptionGroupReq) error {
	o1 := make(map[string]string)
	o2 := make(map[string]string)
	for _, o := range req.Options {
		o1[o.Name] = o.OnValue
		o2[o.Name] = o.OffValue
	}
	if err := validateOptions(o1, req.Advanced, req.Experimental, req.Deprecated); err != nil {
		return err
	}
	return validateOptions(o2, req.Advanced, req.Experimental, req.Deprecated)
}

func optionGroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.OptionGroupReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed)
		return
	}

	if err := validateOptionSet(req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	resp, err := store.Store.Get(context.TODO(), "groupoptions")
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	var groupOptions map[string][]api.VolumeOption
	if err := json.Unmarshal(resp.Kvs[0].Value, &groupOptions); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	var optionSet []api.VolumeOption
	for _, option := range req.Options {
		optionSet = append(optionSet, option)
	}
	groupOptions[req.Name] = optionSet

	groupOptionsJSON, err := json.Marshal(groupOptions)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if _, err := store.Store.Put(context.TODO(), "groupoptions", string(groupOptionsJSON)); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, req)
}
