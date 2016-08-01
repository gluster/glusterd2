package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// APIError is the placeholder for error string to report back to the client
type APIError struct {
	Error string
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
func SendHTTPError(rw http.ResponseWriter, statusCode int, errMsg string) {
	bytes, _ := json.Marshal(APIError{Error: errMsg})
	rw.WriteHeader(statusCode)
	rw.Write(bytes)
}
