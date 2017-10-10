package middleware

import (
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/gorilla/handlers"
)

// LogRequest is a middleware which logs HTTP requests in the
// Apache Common Log Format (CLF)
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := log.WithField("reqid", r.Header.Get("X-Request-ID"))
		handlers.LoggingHandler(l.Writer(), next).ServeHTTP(w, r)
	})
}
