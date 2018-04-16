package events

import (
	"strings"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
)

// New returns a new Event with given information
// Set global to true if event should be broadast across cluster
func New(name string, data map[string]string, global bool) *api.Event {
	return &api.Event{
		ID:        uuid.NewRandom(),
		Name:      strings.ToLower(name),
		Data:      data,
		Global:    global,
		Origin:    gdctx.MyUUID,
		Timestamp: time.Now(),
	}
}

// Broadcast broadcasts events to all registered event handlers
func Broadcast(e *api.Event) error {
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
	registerHooksHandler()
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
