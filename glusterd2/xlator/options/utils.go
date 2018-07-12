package options

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/pkg/utils"
)

// InvalidKeyError is returned by SplitKey if key is not of the correct form.
type InvalidKeyError string

func (e InvalidKeyError) Error() string {
	return fmt.Sprintf("option key not in [<graph>.]<xlator>.<option-name> form: %s", string(e))
}

var validGraphs = [...]string{
	"brick",
	"fuse",
	"gfproxy",
	"nfs",
}

// SplitKey returns three strings by breaking key of the form
// [<graph>].<xlator>.<option-name> into its constituents. <graph> is optional.
// Returns an InvalidKeyError if key is not of correcf form.
func SplitKey(k string) (string, string, string, error) {
	var graph, xlator, optName string

	tmp := strings.Split(strings.TrimSpace(k), ".")
	if len(tmp) < 2 {
		// must at least be of the form <xlator>.<name>
		return graph, xlator, optName, InvalidKeyError(k)
	}

	if utils.StringInSlice(tmp[0], validGraphs[:]) {
		// valid graph present
		if len(tmp) < 3 {
			// must be of the form <graph>.<xlator>.<name>
			return graph, xlator, optName, InvalidKeyError(k)
		}
		graph = tmp[0]
		xlator = tmp[1]
		optName = strings.Join(tmp[2:], ".")
	} else {
		// key is of the format <xlator>.<name> where <name> itself
		// may contain dots. For example: transport.socket.ssl-enabled
		xlator = tmp[0]
		optName = k[len(xlator)+1:]
	}

	return graph, xlator, optName, nil
}
