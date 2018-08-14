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
	for _, o := range req.Options {
		o1[o.Name] = o.OnValue
	}
	return validateOptions(o1, req.Advanced, req.Experimental, req.Deprecated)
}

func optionGroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.OptionGroupReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	if err := validateOptionSet(req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

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

	var optionSet []api.VolumeOption
	for _, option := range req.Options {
		optionSet = append(optionSet, option)
	}

	groupOptions[req.Name] = &api.OptionGroup{
		Name:        req.Name,
		Options:     optionSet,
		Description: req.Description,
	}

	groupOptionsJSON, err := json.Marshal(groupOptions)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if _, err := store.Put(context.TODO(), "groupoptions", string(groupOptionsJSON)); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// FIXME: Change this to http.StatusCreated when we are able to set
	// location header with a unique URL that points to created option
	// group.
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, req)
}
