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
	data := make(map[string]string)
	data["volume.name"] = v.Name
	data["volume.id"] = v.ID.String()

	return events.New(string(e), data, true)
}
