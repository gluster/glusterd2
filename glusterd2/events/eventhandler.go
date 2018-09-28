package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gluster/glusterd2/pkg/api"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

// Handler defines the event handler interface.
// It is registered with the events framework to be called when an event
// happens.
type Handler interface {
	// Handle is the function that gets called when an event occurs.
	// Handle needs to be thread safe, as it can be called concurrently when
	// multiple events arrive at the same time.
	Handle(*api.Event)
	// Events should returns a list of events that the handler is interested in.
	// Return an empty list if interested in all events.
	Events() []string
}

// HandlerID is returned when a Handler is registered. It can be used to unregister a registered Handler.
type HandlerID uint64

// handler implements the Handler interface around a standalone Handle function
type handler struct {
	handle func(*api.Event)
	events []string
}

var (
	handlers struct {
		wg sync.WaitGroup

		sync.RWMutex
		chans map[HandlerID]chan<- *api.Event
		next  HandlerID
	}
)

func init() {
	handlers.chans = make(map[HandlerID]chan<- *api.Event)
}

func addHandler(c chan<- *api.Event) HandlerID {
	handlers.Lock()
	defer handlers.Unlock()

	id := handlers.next
	handlers.chans[id] = c
	handlers.next++

	return id
}

func delHandler(id HandlerID) chan<- *api.Event {
	handlers.Lock()
	defer handlers.Unlock()

	c, ok := handlers.chans[id]
	if !ok {
		return nil
	}
	delete(handlers.chans, id)
	return c
}

// Register a Handler to be called when the events happen.
func Register(h Handler) HandlerID {
	in := make(chan *api.Event)
	id := addHandler(in)

	handlers.wg.Add(1)
	go func() {
		handleEvents(in, h)
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

func handleEvents(in <-chan *api.Event, h Handler) {
	var wg sync.WaitGroup

	events := normalizeEvents(h.Events())

	for e := range in {
		if interested(e, events) {
			wg.Add(1)
			go func(e *api.Event) {
				h.Handle(e)
				wg.Done()
			}(e)
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
func interested(e *api.Event, events []string) bool {
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

// NewHandler returns a Handler wrapping the provided Handle function, and the interested events.
// If no events are provided, the handler is interested in all events.
func NewHandler(handle func(*api.Event), events ...string) Handler {
	return &handler{handle, events}
}

func (h *handler) Handle(e *api.Event) {
	h.handle(e)
}

func (h *handler) Events() []string {
	return h.events
}

func getJWTToken(eventtype, secret string) string {
	claims := &jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Second * 10).Unix(),
		Subject:   eventtype,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return ss
}

//WebhookPublish sends event to registered webhooks
func WebhookPublish(webhook *eventsapi.Webhook, e *api.Event) error {
	message, err := json.Marshal(e)
	if err != nil {
		log.WithError(err).WithField("name", e.Name).Error("failed to marshal event")
		return err
	}

	body := strings.NewReader(string(message))

	req, err := http.NewRequest("POST", webhook.URL, body)
	if err != nil {
		log.WithError(err).Error("error forming the request object")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if webhook.Token != "" {
		req.Header.Set("Authorization", "bearer "+webhook.Token)
	}

	if webhook.Secret != "" {
		token := getJWTToken(e.Name, webhook.Secret)
		req.Header.Set("Authorization", "bearer "+token)
	}

	tr := &http.Transport{
		DisableCompression:    true,
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("error while connecting to webhook")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("webhook responded with status: ", string(resp.StatusCode))
		return fmt.Errorf("failed with unexpected status code %d", resp.StatusCode)
	}
	return nil
}
