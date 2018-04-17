package volume

import (
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/pkg/api"
)

// Event represents Volume life cycle events
type Event string

const (
	// EventVolumeCreated represents Volume Create event
	EventVolumeCreated Event = "volume.created"
	// EventVolumeExpanded represents Volume Expand event
	EventVolumeExpanded = "volume.expanded"
	// EventVolumeStarted represents Volume Start event
	EventVolumeStarted = "volume.started"
	// EventVolumeStopped represents Volume Stop event
	EventVolumeStopped = "volume.stopped"
	// EventVolumeDeleted represents Volume Delete event
	EventVolumeDeleted = "volume.deleted"
)

// NewEvent adds required details to event based on Volume info
func NewEvent(e Event, v *Volinfo) *api.Event {
	data := map[string]string{
		"volume.name": v.Name,
		"volume.id":   v.ID.String(),
	}

	return events.New(string(e), data, true)
}
