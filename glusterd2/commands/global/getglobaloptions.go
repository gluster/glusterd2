package globalcommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/cluster"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

func getGlobalOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	c, err := cluster.GetCluster()
	// ErrClusterNotFound here implies that no global option has yet been explicitly set. Ignoring it.
	if err != nil && err != errors.ErrClusterNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("Problem retrieving cluster information from etcd store: %s", err.Error()))
		return
	}

	resp := createGlobalOptionsGetResp(c)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createGlobalOptionsGetResp(c *cluster.Cluster) []api.GlobalOptionsGetResp {
	var resp []api.GlobalOptionsGetResp

	for k, v := range cluster.GlobalOptMap {
		var val string
		var found bool
		if c == nil {
			val = v.DefaultValue
			found = false
		} else {
			val, found = c.Options[k]
			if !found {
				val = v.DefaultValue
			}
		}

		resp = append(resp, api.GlobalOptionsGetResp{
			Key:          k,
			Value:        val,
			DefaultValue: v.DefaultValue,
			Modified:     found,
		})
	}

	return resp
}
