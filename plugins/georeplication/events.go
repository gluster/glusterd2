package georeplication

import (
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/pkg/api"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
)

type georepEvent string

const (
	eventGeorepCreated     georepEvent = "georep.created"
	eventGeorepStarted                 = "georep.started"
	eventGeorepStopped                 = "georep.stopped"
	eventGeorepDeleted                 = "georep.deleted"
	eventGeorepPaused                  = "georep.paused"
	eventGeorepResumed                 = "georep.resumed"
	eventGeorepConfigSet               = "georep.config.set"
	eventGeorepConfigReset             = "georep.config.reset"
)

func newGeorepEvent(e georepEvent, session *georepapi.GeorepSession, extra *map[string]string) *api.Event {
	data := make(map[string]string)

	if session != nil {
		data = map[string]string{
			"master.name":   session.MasterVol,
			"master.id":     session.MasterID.String(),
			"remote.name":   session.RemoteVol,
			"remote.id":     session.RemoteID.String(),
			"remote.host":   session.RemoteHosts[0].Hostname,
			"remote.peerid": session.RemoteHosts[0].PeerID.String(),
			"remote.user":   session.RemoteUser,
		}
	}

	if extra != nil {
		for k, v := range *extra {
			data[k] = v
		}
	}

	return events.New(string(e), data, true)
}
