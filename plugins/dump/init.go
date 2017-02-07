package dumpplugin

import (
	"github.com/gluster/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/servers/sunrpc/program"
)


type DumpPlugin struct{
}

func (p *DumpPlugin) SunRpcProcedures() program.Program {
	return &GfDump{}
}

func (p *DumpPlugin) RestRoutes() route.Routes {
	return nil
}

func (p *DumpPlugin) RegisterStepFuncs() {
	return
}
