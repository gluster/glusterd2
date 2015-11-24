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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if e := json.NewEncoder(w).Encode(rsp); e != nil {
		log.WithField("error", e).Error("Failed to send the response -", rsp)
	}
	return
}

// SendHTTPError is to report error back to the client
func SendHTTPError(rw http.ResponseWriter, statusCode int, errMsg string) {
	bytes, _ := json.Marshal(APIError{Error: errMsg})
	rw.WriteHeader(statusCode)
	rw.Write(bytes)
}
