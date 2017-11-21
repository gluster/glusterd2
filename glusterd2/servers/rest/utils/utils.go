// Package utils provides utility functions for working with the GD2 rest server
package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
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

// SendHTTPResponse to send response back to the client
func SendHTTPResponse(w http.ResponseWriter, statusCode int, rsp interface{}) {
	if rsp != nil {
		// Do not include content-type header for responses such as 204
		// which as per RFC, should not have a response body.
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	}
	// Maintain the order of these calls that modify http.ResponseWriter
	// object.
	w.WriteHeader(statusCode)
	if rsp != nil {
		if e := json.NewEncoder(w).Encode(rsp); e != nil {
			log.WithField("error", e).Error("Failed to send the response -", rsp)
		}
	}
	return
}

// SendHTTPError is to report error back to the client
func SendHTTPError(rw http.ResponseWriter, statusCode int, errMsg string, errCode api.ErrorCode) {
	bytes, _ := json.Marshal(APIError{Code: errCode, Error: errMsg})
	rw.WriteHeader(statusCode)
	rw.Write(bytes)
}

// GetReqIDandLogger returns a request ID and a request-scoped logger having
// the request ID as a logging field.
func GetReqIDandLogger(r *http.Request) (string, *log.Entry) {
	reqID := r.Header.Get("X-Request-ID")
	return reqID, log.WithField("reqid", reqID)
}
