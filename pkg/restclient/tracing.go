package restclient

import (
	"net/http"

	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"
)

// TraceEnable enables tracing
func (c *Client) TraceEnable(req tracemgmtapi.SetupTracingReq) (tracemgmtapi.JaegerConfigInfo, error) {
	var jaegercfginfo tracemgmtapi.JaegerConfigInfo
	err := c.post("/v1/tracemgmt", req, http.StatusCreated, &jaegercfginfo)
	return jaegercfginfo, err
}
