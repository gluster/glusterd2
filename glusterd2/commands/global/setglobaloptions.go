package globalcommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/cluster"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

func setGlobalOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.GlobalOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	c, err := cluster.GetCluster()
	// ErrClusterNotFound here implies that no global option has yet been explicitly set. Ignoring it.
	if err != nil && err != errors.ErrClusterNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("Problem retrieving cluster information from etcd store: %s", err.Error()))
		return
	}

	for k, v := range req.Options {
		if _, found := cluster.GlobalOptMap[k]; found {
			// TODO validate the value type for global option

			if c == nil {
				c = new(cluster.Cluster)
				c.Options = make(map[string]string)
			}
			c.Options[k] = v
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("Invalid global option: %s", k))
			continue
		}
	}

	if err := cluster.UpdateCluster(c); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to update store with cluster attributes %s", err.Error()))
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, c.Options)
}
