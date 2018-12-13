package xlator

import (
	"github.com/gluster/glusterd2/glusterd2/options"
)

// Xlator represents a GlusterFS xlator
type Xlator struct {
	ID        string
	Options   []*options.Option
	Flags     uint32
	OpVersion []uint32

	// Not loaded from .so, set by glusterd2 code
	Validate ValidationFunc
	Actor    OptionActor

	// This is pretty much useless now.
	rawID uint32

	// Category is parent directory name
	Category string
}

// FullName returns xlator name including the category name
func (xl *Xlator) FullName() string {
	return xl.Category + "/" + xl.ID
}
