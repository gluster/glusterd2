// The command package defines the command interfaces that need to be implemented by the GlusterD commands
package command

import (
	"github.com/gorilla/mux"
)

type Command interface {
	// SetRoutes should setup the REST API endpoints and handlers for the command
	SetRoutes(r *mux.Router) error
	// SetTransactionHandlers will setup the transaction handlers for the command
	SetTransactionHandlers() error
}
