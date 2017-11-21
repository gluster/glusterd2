package middleware

import (
	"net/http"

	"github.com/pborman/uuid"
)

// TODO
// In Go, the idiomatic and recommended way to attach any request scoped
// metadata information across goroutine and process boundaries is to use the
// 'context' package. This is not useful unless we pass down this context
// all through-out the request scope across nodes, and this involves some
// code changes in function signatures at many places
// The following simple implementation is good enough until then...

// ReqIDGenerator is a middleware which generates a UUID for each incoming
// HTTP request and sets this UUID as a header in request and in response.
func ReqIDGenerator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the request id sent by the client, if any
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewRandom().String()
			r.Header.Set("X-Request-ID", reqID)
		}
		// Set response header
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r)
	})
}
