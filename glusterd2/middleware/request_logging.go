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
		f := log.Fields{
			"subsys": "rest",
			"reqid":  gdctx.GetReqID(r.Context()).String(),
		}
		if origin := r.Header.Get("X-Gluster-Origin-Args"); origin != "" {
			f["origin"] = origin
		}
		entry := log.WithFields(f)
		entry.WithFields(log.Fields{
			"method": r.Method,
			"url":    r.URL,
			"client": r.UserAgent(),
		}).Info("HTTP Request")
		delete(entry.Data, logging.SourceField)
		handlers.CombinedLoggingHandler(entry.Writer(), next).ServeHTTP(w, r)
	})
}
