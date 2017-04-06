package helloplugin

import (
	"github.com/gluster/glusterd2/servers/rest/route"
	"github.com/prashanthpai/sunrpc"
)

type HelloPlugin struct {
}

func (p *HelloPlugin) SunRpcProgram() sunrpc.Program {
	return nil
}

func (p *HelloPlugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "HelloGet",
			Method:      "GET",
			Pattern:     "/hello",
			Version:     1,
			HandlerFunc: helloGetHandler},
		route.Route{
			Name:        "HelloPost",
			Method:      "POST",
			Pattern:     "/hello",
			Version:     1,
			HandlerFunc: helloPostHandler},
	}
}

func (p *HelloPlugin) RegisterStepFuncs() {
	return
}
