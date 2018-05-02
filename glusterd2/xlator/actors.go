package xlator

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	log "github.com/sirupsen/logrus"
)

// OptionActor is an interface that contains Do and Undo methods. These methods
// are invoked during volume set on ALL nodes that make up the volume.
// Each plugin or xlator can provide a type satisying OptionActor interface to
// have the xlator/feature specific logic executed during volume set. An example
// of such logic is the task of starting and stopping daemon.
type OptionActor interface {
	// Do function takes volinfo, option key, option value.
	Do(*volume.Volinfo, string, string) error
	// Undo function takes volinfo, option key, option value. The returned
	// error is currently ignored.
	Undo(*volume.Volinfo, string, string) error
}

// RegisterOptionActor registers a xlator's type implementing OptionActor
// interface. The Do() and Undo() methods of the interface will be invoked
// later during volume set operation.
func RegisterOptionActor(xlator string, actor OptionActor) error {
	xl, err := Find(xlator)
	if err != nil {
		log.WithError(err).WithField("xlator",
			xlator).Error("Could not register xlator actor type")
		return err
	}
	xl.Actor = actor
	return nil
}
