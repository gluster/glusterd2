package globalcommands

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/cluster"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"
)

func getGlobalOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	optname := mux.Vars(r)["optname"]

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

	response := createGlobalOptionsGetResp(c, optname)

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, response)
}

func createGlobalOptionsGetResp(c cluster.Cluster, optname string) *api.GlobalOptionsGetResp {
	return &api.GlobalOptionsGetResp{
		Key:   optname,
		Value: c.Options[optname],
	}
}
