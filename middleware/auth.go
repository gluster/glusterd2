package middleware

import (
	"net/http"
	"strings"
)

func isAuthorized(r *http.Request) bool {
	// TODO: Add configuration to enable this Auth
	// This configuration should not be as command line option
	// Because, one glusterd can be run with `--no-auth` and
	// other glusterd with auth enabled.
	if strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
		return true
	}
	// TODO: Add Auth for requests from Remote IPs
	// Without that, this feature is no impact on
	// existing(Without Auth). Returning true so that it
	// will not break the existing consumers/tests till Auth introduced
	return true
}

// Auth is a middleware which validates the incoming request is from local node or remote node
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthorized(r) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}
