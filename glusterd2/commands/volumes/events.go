package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/volume"
)

type volumeEvent string

const (
	eventVolumeCreated volumeEvent = "volume.created"
	eventVolumeStarted             = "volume.started"
	eventVolumeStopped             = "volume.stopped"
	eventVolumeDeleted             = "volume.deleted"
)

func newVolumeEvent(e volumeEvent, v *volume.Volinfo) *events.Event {
	data := map[string]string{
		"volume.name": v.Name,
		"volume.id":   v.ID.String(),
	}

	return events.New(string(e), data, true)
}
