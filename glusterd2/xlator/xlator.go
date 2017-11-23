package xlator

import (
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
)

// Xlator represents a GlusterFS xlator
type Xlator struct {
	ID      string
	Options []*options.Option
}
