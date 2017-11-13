package plugins

import (
	"github.com/gluster/glusterd2/bin/glusterd2/servers/rest/route"
	"github.com/prashanthpai/sunrpc"
)

// GlusterdPlugin is an interface that every Glusterd plugin will
// implement to add sunrpc program, REST routes and Transaction step
// functions
type GlusterdPlugin interface {
	Name() string
	SunRPCProgram() sunrpc.Program
	RestRoutes() route.Routes
	RegisterStepFuncs()
}
