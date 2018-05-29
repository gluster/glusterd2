// +build plugins

package plugin

import (
	"github.com/gluster/glusterd2/plugins/bitrot"
	"github.com/gluster/glusterd2/plugins/device"
	"github.com/gluster/glusterd2/plugins/events"
	"github.com/gluster/glusterd2/plugins/georeplication"
	"github.com/gluster/glusterd2/plugins/glustershd"
	"github.com/gluster/glusterd2/plugins/quota"
	"github.com/gluster/glusterd2/plugins/rebalance"

	// ensure init() of non-plugins also gets executed
	_ "github.com/gluster/glusterd2/plugins/afr"
	_ "github.com/gluster/glusterd2/plugins/dht"
)

// PluginsList is a list of plugins which implements GlusterdPlugin interface
var PluginsList = []GlusterdPlugin{
	&georeplication.Plugin{},
	&bitrot.Plugin{},
	&quota.Plugin{},
	&events.Plugin{},
	&glustershd.Plugin{},
	&device.Plugin{},
	&rebalance.Plugin{},
}
