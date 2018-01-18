// +build plugins

package plugin

import (
	"github.com/gluster/glusterd2/plugins/georeplication"
	"github.com/gluster/glusterd2/plugins/hello"
	"github.com/gluster/glusterd2/plugins/rebalance"
)

// PluginsList is a list of plugins which implements GlusterdPlugin interface
var PluginsList = []GlusterdPlugin{
	&hello.Plugin{},
	&georeplication.Plugin{},
	&rebalance.Plugin{},
}
