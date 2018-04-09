package api

import (
	"time"

	"github.com/pborman/uuid"
)

// Event represents an event in GD2
type Event struct {
	// ID is a unique event ID
	ID uuid.UUID `json:"id"`
	// Name is the the name of the event
	Name string `json:"name"`
	// Data is any additional data attached to the event.
	Data map[string]string `json:"data,omitempty"`
	// global should be set to true to broadcast event to the full GD2 cluster.
	// If not event is only broadcast in the local node
	Global bool `json:"-"`
	// Origin is used when broadcasting global events to prevent origin nodes
	// rebroadcasting a global event. Event generators need not set this.
	Origin uuid.UUID `json:"origin"`
	// Timestamp is the time when the event was created
	Timestamp time.Time `json:"timestamp"`
}
