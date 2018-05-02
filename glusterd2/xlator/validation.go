package xlator

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	log "github.com/sirupsen/logrus"
)

// ValidationFunc is a function that is invoked during volume set. Each plugin
// or xlator can provide such validation function.
type ValidationFunc func(*volume.Volinfo, string, string) error

// RegisterValidationFunc registers a xlator's validation function for calling
// later during volume set operation.
func RegisterValidationFunc(xlator string, vf ValidationFunc) error {
	xl, err := Find(xlator)
	if err != nil {
		log.WithError(err).WithField("xlator",
			xlator).Error("Could not register xlator validation function")
		return err
	}
	xl.Validate = vf
	return nil
}
