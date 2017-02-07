package plugins

import (
	"github.com/gluster/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/servers/sunrpc/program"
)

type Gd2Plugin interface{
	SunRpcProcedures() program.Program
	RestRoutes() route.Routes
	RegisterStepFuncs()
}
