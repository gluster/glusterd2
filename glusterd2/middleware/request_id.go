package middleware

import (
	"context"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// ReqIDGenerator is a middleware which generates a UUID for each incoming
// HTTP request and sets this UUID as a header in request and in response.
func ReqIDGenerator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Generate request ID and set it in request context
		reqID := uuid.NewRandom().String()
		ctx := context.WithValue(r.Context(), gdctx.ReqIDKey, reqID)

		// Also set request ID in request and response headers
		w.Header().Set("X-Request-ID", reqID)

		// Create request-scoped logger and set in request context
		reqLogger := log.WithField("reqid", reqID)
		ctx = context.WithValue(ctx, gdctx.ReqLoggerKey, reqLogger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
