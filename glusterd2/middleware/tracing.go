package middleware

import (
	"net/http"

	"go.opencensus.io/plugin/ochttp"
)

// Tracing is a http middleware to be use for trace incoming http.Request
func Tracing(next http.Handler) http.Handler {
	return &ochttp.Handler{Handler: next}
}
