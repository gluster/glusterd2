package xlator

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	log "github.com/sirupsen/logrus"
)

var optionActors = make(map[string]OptionActor)

//GetOptActors returns a mapping of xlator ID to its action interface implementation
func GetOptActors() map[string]OptionActor {
	return optionActors
}

//VolumeOpType type associated with volume operation
type VolumeOpType uint8

//constants definded to represent volume operation types
const (
	VolumeSet VolumeOpType = iota
	VolumeReset
	VolumeStart
	VolumeStop
)

// OptionActor is an interface that contains Do and Undo methods. These methods
// are invoked during volume set on ALL nodes that make up the volume.
// Each plugin or xlator can provide a type satisying OptionActor interface to
// have the xlator/feature specific logic executed during volume set. An example
// of such logic is the task of starting and stopping daemon.
type OptionActor interface {
	// Do function takes volinfo, option key, option value,VolumeOpType, logger.
	Do(*volume.Volinfo, string, string, VolumeOpType, log.FieldLogger) error
	// Undo function takes volinfo, option key, option value,VolumeOpType, logger. The returned
	// error is currently ignored.
	Undo(*volume.Volinfo, string, string, VolumeOpType, log.FieldLogger) error
}

// RegisterOptionActor registers a xlator's type implementing OptionActor
// interface. The Do() and Undo() methods of the interface will be invoked
// later during volume set operation.
func RegisterOptionActor(xlator string, actor OptionActor) {
	optionActors[xlator] = actor
}
