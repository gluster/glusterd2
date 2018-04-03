package peercommands

import (
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
)

type peerEvent string

const (
	eventPeerAdded   peerEvent = "peer.added"
	eventPeerRemoved           = "peer.removed"
)

func newPeerEvent(e peerEvent, p *peer.Peer) *api.Event {
	data := map[string]string{
		"peer.id":   p.ID.String(),
		"peer.name": p.Name,
	}

	return events.New(string(e), data, true)
}
