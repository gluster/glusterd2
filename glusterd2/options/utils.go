package options

import (
	"path"
	"strings"

	"github.com/gluster/glusterd2/pkg/utils"
)

// SplitKey returns three strings by breaking key of the form
// [<graph>].<xlator>.<option-name> into its constituents. <graph> is optional.
func SplitKey(k string) (string, string, string) {
	var graph, xlator, optName string

	tmp := strings.Split(strings.TrimSpace(k), ".")
	if len(tmp) == 1 {
		// If only xlator name is specified to enable/disable that xlator

		// Remove category prefix for example "cluster/replicate"
		// will be converted to "replicate"
		optName = path.Base(tmp[0])

		return graph, tmp[0], optName
	}

	if utils.StringInSlice(tmp[0], utils.ValidVolfiles[:]) {
		// valid graph present
		graph = tmp[0]
		xlator = tmp[1]

		if len(tmp) < 3 {
			// may be only <graph>.<xlator> specified

			// Remove category prefix for example "cluster/replicate"
			// will be converted to "replicate"
			optName = path.Base(xlator)

			return graph, xlator, optName
		}
		optName = strings.Join(tmp[2:], ".")
	} else {
		// key is of the format <xlator>.<name> where <name> itself
		// may contain dots. For example: transport.socket.ssl-enabled
		xlator = tmp[0]
		optName = k[len(xlator)+1:]
	}

	return graph, xlator, optName
}
