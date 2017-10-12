// +build plugins

package plugins

import (
	"github.com/gluster/glusterd2/plugins/hello"
	"github.com/gluster/glusterd2/plugins/snapshot"
)

// PluginsList is a list of plugins which implements GlusterdPlugin interface
var PluginsList = []GlusterdPlugin{
	&hello.Plugin{},
	&snapshot.Plugin{},
}
