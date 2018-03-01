package events

import (
	"strings"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

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
	global bool
	// Origin is used when broadcasting global events to prevent origin nodes
	// rebroadcasting a global event. Event generators need not set this.
	Origin uuid.UUID `json:"origin"`
	// Timestamp is the time when the event was created
	Timestamp time.Time `json:"timestamp"`
}

// New returns a new Event with given information
// Set global to true if event should be broadast across cluster
func New(name string, data map[string]string, global bool) *Event {
	return &Event{
		ID:        uuid.NewRandom(),
		Name:      strings.ToLower(name),
		Data:      data,
		global:    global,
		Origin:    gdctx.MyUUID,
		Timestamp: time.Now(),
	}
}

// Broadcast broadcasts events to all registered event handlers
func Broadcast(e *Event) error {
	handlers.RLock()
	defer handlers.RUnlock()

	for _, h := range handlers.chans {
		h <- e
	}

	return nil
}

// Start starts the events framework.
// Should be called only after store is up.
func Start() error {
	StartGlobal()
	startEventLogger()
	registerGaneshaHandler()
	startLivenessWatcher()
	return nil
}

// Stop stops the events framework, events will no longer be broadcast
func Stop() error {
	stopLivenessWatcher()
	stopEventLogger()
	StopGlobal()
	stopHandlers()

	return nil
}
