package volumecommands

import (
	"github.com/gluster/glusterd2/pkg/api"
)

// GroupOptions maps from a profile name to a set of options
var defaultGroupOptions = map[string][]api.Option{
	"profile.test": {{"afr.eager-lock", "on", "off"},
		{"gfproxy.afr.eager-lock", "on", "off"}},
}
