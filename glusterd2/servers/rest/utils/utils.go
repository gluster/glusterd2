// Package utils provides utility functions for working with the GD2 rest server
package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/api"
)

// APIError is the placeholder for error string to report back to the client
type APIError struct {
	Code  api.ErrorCode `json:"error_code"`
	Error string        `json:"error"`
}

// UnmarshalRequest unmarshals JSON in `r` into `v`
func UnmarshalRequest(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// SendHTTPResponse sends non-error response to the client.
func SendHTTPResponse(ctx context.Context, w http.ResponseWriter, statusCode int, resp interface{}) {

	if resp != nil {
		// Do not include content-type header for responses such as 204
		// which as per RFC, should not have a response body.
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	}

	w.Header().Set("X-Gluster-Node-Id", gdctx.MyUUID.String())
	w.Header().Set("X-Gluster-Cluster-Id", gdctx.MyClusterID.String())

	w.WriteHeader(statusCode)

	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger := gdctx.GetReqLogger(ctx)
			logger.WithError(err).Error("Failed to send the response -", resp)
		}
	}
}

// SendHTTPError sends an error response to the client.
func SendHTTPError(ctx context.Context, w http.ResponseWriter, statusCode int, errMsg string, errCode api.ErrorCode) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("X-Gluster-Node-Id", gdctx.MyUUID.String())
	w.Header().Set("X-Gluster-Cluster-Id", gdctx.MyClusterID.String())

	w.WriteHeader(statusCode)

	resp := APIError{Code: errCode, Error: errMsg}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger := gdctx.GetReqLogger(ctx)
		logger.WithError(err).Error("Failed to send the response -", resp)
	}
}
