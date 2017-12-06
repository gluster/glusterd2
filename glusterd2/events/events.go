package events

import (
	"github.com/pborman/uuid"
)

// Event represents an event in GD2
type Event struct {
	// ID is a unique event ID
	ID uuid.UUID
	// Name is the the name of the event
	Name string
	// Data is any additional data attached to the event. Should be a JSON
	// document
	Data string
	// global should be set to true to broadcast event to the full GD2 cluster.
	// If not event is only broadcast in the local node
	global bool
}

// New returns a new Event with given information
// Set global to true if event should be broadast across cluster
func New(name, data string, global bool) Event {
	return Event{
		ID:     uuid.NewRandom(),
		Name:   name,
		Data:   data,
		global: global,
	}
}

// Broadcast broadcasts events to all registered event handlers
func Broadcast(e Event) error {
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
	// Nothing much to start other than global stuff
	startGlobal()

	return nil
}

// Stop stops the events framework, events will no longer be broadcast
func Stop() error {
	stopGlobal()
	stopHandlers()

	return nil
}
