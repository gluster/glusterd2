package optionscommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/options"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
)

const (
	lockKey = "clusteroptions"
)

func setClusterOptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.ClusterOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	txn, err := oldtransaction.NewTxnWithLocks(ctx, lockKey)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	c, err := options.GetClusterOptions()
	if err != nil && err != errors.ErrClusterOptionsNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if c == nil {
		c = new(options.ClusterOptions)
		c.Options = make(map[string]string)
	}

	for k, v := range req.Options {
		if opt, found := options.ClusterOptMap[k]; found {
			if opt.ValidateFunc != nil {
				if err := opt.ValidateFunc(k, v); err != nil {
					restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
						fmt.Sprintf("%s failed validation: %s", k, err))
					return
				}
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
