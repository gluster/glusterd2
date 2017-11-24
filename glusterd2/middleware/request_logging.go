package middleware

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gorilla/handlers"
)

// LogRequest is a middleware which logs HTTP requests in the
// Apache Common Log Format (CLF)
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := gdctx.GetReqLogger(r.Context())
		handlers.LoggingHandler(l.Writer(), next).ServeHTTP(w, r)
	})
}
