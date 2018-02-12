package plugin

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/sunrpc"
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
