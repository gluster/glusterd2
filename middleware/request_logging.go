package middleware

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
)

// LogRequest is a middleware which logs HTTP requests in the
// Apache Common Log Format (CLF)
func LogRequest(next http.Handler) http.Handler {
	return handlers.LoggingHandler(log.StandardLogger().Writer(), next)
}
