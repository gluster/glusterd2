package middleware

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/logging"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

// LogRequest is a middleware which logs HTTP requests in the
// Apache Common Log Format (CLF)
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := log.WithField("reqid", gdctx.GetReqID(r.Context()).String())
		delete(l.Data, logging.SourceField)
		handlers.LoggingHandler(l.Writer(), next).ServeHTTP(w, r)
	})
}
