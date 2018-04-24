package plugin

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
)

// GlusterdPlugin is an interface that every Glusterd plugin will
// implement to add REST routes and Transaction step
// functions
type GlusterdPlugin interface {
	Name() string
	RestRoutes() route.Routes
	RegisterStepFuncs()
}
