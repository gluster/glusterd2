package plugins

import (
	"github.com/gluster/glusterd2/servers/rest/route"
	"github.com/prashanthpai/sunrpc"
)

type Gd2Plugin interface {
	SunRpcProgram() sunrpc.Program
	RestRoutes() route.Routes
	RegisterStepFuncs()
}
