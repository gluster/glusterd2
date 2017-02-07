package plugins

import (
	"github.com/gluster/glusterd2/plugins/dump"
	"github.com/gluster/glusterd2/plugins/hello"
)

var PluginsList = []Gd2Plugin{
	&dumpplugin.DumpPlugin{},
	&helloplugin.HelloPlugin{},
}
