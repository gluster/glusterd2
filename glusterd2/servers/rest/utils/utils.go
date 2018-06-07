// Package utils provides utility functions for working with the GD2 rest server
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
)

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

	w.WriteHeader(statusCode)

	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger := gdctx.GetReqLogger(ctx)
			logger.WithError(err).Error("Failed to send the response -", resp)
		}
	}
}

// SendHTTPError sends an error response to the client. The caller of this
// function can pass either the error or one or more error code(s) exported by
// api package. Example usage:
// SendHTTPError(ctx, http.StatusBadRequest, err) // Pass error as is
// SendHTTPError(ctx, http.StatusBadRequest, "", api.ErrorCode) // Specify error code
// SendHTTPError(ctx, http.StatusBadRequest, "custom error") // Pass specific error string
func SendHTTPError(ctx context.Context, w http.ResponseWriter, statusCode int,
	err interface{}, errCodes ...api.ErrorCode) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var resp api.ErrorResp
	if v, ok := err.(api.ErrorResponse); ok {
		// if the passed error type implements the api.ErrorResponse
		// interface, we don't have aything else to do
		resp = v.Response()
		statusCode = v.Status()
	} else {
		errMsg := fmt.Sprint(err)
		if errMsg != "" && errMsg != "<nil>" || len(errCodes) == 0 {
			resp.Errors = append(resp.Errors, api.HTTPError{
				Code:    int(api.ErrCodeGeneric),
				Message: errMsg})
		} else {
			for _, code := range errCodes {
				resp.Errors = append(resp.Errors, api.HTTPError{
					Code:    int(code),
					Message: api.ErrorCodeMap[code]})
			}
		}
	}

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger := gdctx.GetReqLogger(ctx)
		logger.WithError(err).Error("Failed to send the response -", resp)
	}
}

// ErrToStatusCode returns error and http Status code based on  err
func ErrToStatusCode(err error) (int, error) {
	var statuscode int

	switch err {
	case gderrors.ErrPeerNotFound:
		statuscode = http.StatusNotFound
	case gderrors.ErrVolNotFound:
		statuscode = http.StatusNotFound
	case gderrors.ErrSnapNotFound:
		statuscode = http.StatusNotFound
	case transaction.ErrLockTimeout:
		statuscode = http.StatusConflict
	default:
		statuscode = http.StatusInternalServerError
	}
	return statuscode, err
}
