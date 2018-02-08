package globalcommands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/cluster"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

func setGlobalOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.GlobalOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	var c cluster.Cluster
	resp, err := store.Store.Get(context.TODO(), cluster.ClusterPrefix)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrClusterNotFound.Error(), api.ErrCodeDefault)
		return
	}

	if resp.Count != 1 {
		return
	}

	if err = json.Unmarshal(resp.Kvs[0].Value, &c); err != nil {
		return
	}

	for k, v := range req.Options {
		if _, found := cluster.GlobalOptMap[k]; found {
			// TODO validate the value type for global option

			c.Options[k] = v
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("Invalid global option: %s", k), api.ErrCodeDefault)
			continue
		}
	}

	data, _ := json.Marshal(c)
	if _, err := store.Store.Put(context.TODO(), cluster.ClusterPrefix, string(data)); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprint("Failed to update store with cluster attributes %s: %s", err.Error()), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, c.Options)
}
