package optionscommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/options"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

func setClusterOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.ClusterOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	// TODO: Take lock here

	c, err := options.GetClusterOptions()
	if err != nil && err != errors.ErrClusterOptionsNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	for k, v := range req.Options {
		if _, found := options.ClusterOptMap[k]; found {
			// TODO validate the value type for global option

			if c == nil {
				c = new(options.ClusterOptions)
				c.Options = make(map[string]string)
			}
			c.Options[k] = v
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("Invalid global option: %s", k))
			return
		}
	}

	if err := options.UpdateClusterOptions(c); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to update store with cluster attributes %s", err.Error()))
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, c.Options)
}
