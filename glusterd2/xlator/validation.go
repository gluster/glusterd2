package xlator

import (
	"github.com/gluster/glusterd2/glusterd2/volume"
)

var validationFuncs = make(map[string]ValidationFunc)

// ValidationFunc is a function that is invoked during volume set. Each plugin
// or xlator can provide such validation function.
type ValidationFunc func(*volume.Volinfo, string, string) error

// RegisterValidationFunc registers a xlator's validation function for calling
// later during volume set operation.
func RegisterValidationFunc(xlator string, vf ValidationFunc) {
	validationFuncs[xlator] = vf
}
