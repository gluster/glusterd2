package middleware

import (
	"expvar"
	"net/http"
)

var expReqs = expvar.NewMap("http")

type statusRecorder struct {
	http.ResponseWriter
}

func (rec *statusRecorder) WriteHeader(code int) {
	expReqs.Add("total_requests_received", 1)
	switch code / 100 {
	case 2:
		expReqs.Add("success", 1)
	case 4:
		expReqs.Add("client_errors", 1)
	case 5:
		expReqs.Add("server_errors", 1)
	}
	rec.ResponseWriter.WriteHeader(code)
}

// Expvar is a middleware which updates some metrics about requests
func Expvar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expReqs.Add("in_progress", 1)
		next.ServeHTTP(&statusRecorder{w}, r)
		expReqs.Add("in_progress", -1)
	})
}
