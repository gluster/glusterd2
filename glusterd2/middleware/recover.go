package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// Recover will recover() from unexpected panics in the code and ensures
// that clients get a 500 error response. The stack is logged during panic.
func Recover(next http.Handler) http.Handler {
	return handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(next)
}
