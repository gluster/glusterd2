// +build plugins

package plugin

import (
	"github.com/gluster/glusterd2/plugins/georeplication"
	"github.com/gluster/glusterd2/plugins/glustershd"
	"github.com/gluster/glusterd2/plugins/hello"
)

// PluginsList is a list of plugins which implements GlusterdPlugin interface
var PluginsList = []GlusterdPlugin{
	&hello.Plugin{},
	&georeplication.Plugin{},
	&glustershd.Plugin{},
}
