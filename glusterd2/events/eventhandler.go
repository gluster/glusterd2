package events

import (
	"sort"
	"strings"
	"sync"
)

// Handler is a function that is registered to be called when an event happens
// event is the name of the event that happened and data is any extra data that
// was attached to the event as a json document
type Handler func(*Event)

// HandlerID is returned when a Handler is registered. It can be used to unregister a registered Handler.
type HandlerID uint64

var (
	handlers struct {
		wg sync.WaitGroup

		sync.RWMutex
		chans map[HandlerID]chan<- *Event
		next  HandlerID
	}
)

func init() {
	handlers.chans = make(map[HandlerID]chan<- *Event)
}

func addHandler(c chan<- *Event) HandlerID {
	handlers.Lock()
	defer handlers.Unlock()

	id := handlers.next
	handlers.chans[id] = c
	handlers.next++

	return id
}

func delHandler(id HandlerID) chan<- *Event {
	handlers.Lock()
	defer handlers.Unlock()

	c, ok := handlers.chans[id]
	if !ok {
		return nil
	}
	delete(handlers.chans, id)
	return c
}

// Register a Handler to be called when the given events happen.
// If no events are specified, handler function is called for all events
// Handlers need to be thread safe, as they can be called concurrently when
// multiple events arrive at the same time.
func Register(h Handler, events ...string) HandlerID {
	in := make(chan *Event)
	id := addHandler(in)

	go func() {
		handlers.wg.Add(1)
		handleEvents(in, h, events...)
		handlers.wg.Done()
	}()

	return id
}

// Unregister stops a registered Handler from being called for any further
// events
func Unregister(id HandlerID) {
	c := delHandler(id)
	if c != nil {
		close(c)
	}
}

func handleEvents(in <-chan *Event, h Handler, events ...string) {
	var wg sync.WaitGroup

	events = normalizeEvents(events)

	for e := range in {
		if interested(e, events) {
			go func() {
				wg.Add(1)
				h(e)
				wg.Done()
			}()
		}
	}

	wg.Wait()
}

// normalizeEvents normalizes given list to lower case and then sorts it
func normalizeEvents(events []string) []string {
	for i, v := range events {
		events[i] = strings.ToLower(v)
	}
	sort.Strings(events)
	return events
}

// interested returns true if given event is found in the events list
// Returns true if found or if list is empty
func interested(e *Event, events []string) bool {
	if len(events) == 0 {
		return true
	}
	i := sort.SearchStrings(events, e.Name)
	return events[i] == e.Name
}

// stopHandlers stops all registered handlers
func stopHandlers() error {
	handlers.Lock()
	defer handlers.Unlock()

	for id, ch := range handlers.chans {
		delete(handlers.chans, id)
		close(ch)
	}
	handlers.wg.Wait()

	return nil
}
