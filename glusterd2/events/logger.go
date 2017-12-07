package events

import (
	log "github.com/sirupsen/logrus"
)

var lhID HandlerID

// eventLogger logs all events being generated in the GD2 cluster
// The events are logged in DEBUG level only.
func eventLogger(e *Event) {
	log.WithFields(log.Fields{
		"event.id":   e.ID.String(),
		"event.name": e.Name,
	}).Debug("new event")
}

// startEventLogger registers the eventLogger with the events framework
func startEventLogger() {
	lhID = Register(eventLogger)
}

// stopEventLogger unregisters the eventLogger
func stopEventLogger() {
	Unregister(lhID)
}
