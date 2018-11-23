package optionscommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/options"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

func getClusterOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	c, err := options.GetClusterOptions()
	if err != nil && err != errors.ErrClusterOptionsNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := createClusterOptionsResp(c)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createClusterOptionsResp(c *options.ClusterOptions) []api.ClusterOptionsResp {
	var resp []api.ClusterOptionsResp

	for k, v := range options.ClusterOptMap {
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

		resp = append(resp, api.ClusterOptionsResp{
			Key:          k,
			Value:        val,
			DefaultValue: v.DefaultValue,
			Modified:     found,
		})
	}

	return resp
}
