package plugins

import (
	"github.com/gluster/glusterd2/plugins/hello"
)

var PluginsList = []Gd2Plugin{
	&helloplugin.HelloPlugin{},
}
