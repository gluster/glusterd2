package events

import (
	"encoding/json"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const (
	eventsPrefix           = "events/"
	defaultEventsTTL int64 = 600
)

var (
	// globalHandler id
	ghID HandlerID
	// globalListener wait group and stop channel
	glWg   sync.WaitGroup
	glStop chan struct{}
)

// globalHandler listens for events that are global and broadcasts them across the cluster
func globalHandler(ev *api.Event) {
	if !ev.Global {
		return
	}

	k := eventsPrefix + ev.ID.String()
	ev.Origin = gdctx.MyUUID
	v, err := json.Marshal(ev)
	if err != nil {
		log.WithFields(log.Fields{
			"event.id":   ev.ID.String(),
			"event.name": ev.Name,
		}).WithError(err).Error("failed global broadcast, failed to marshal event")
	}

	// Putting event with a TTL so that we don't have stale events lingering in store
	// Using a TTL of 10 minutes(configurable) should allow all members in the
	// cluster to receive event
	eventsttl := config.GetInt64("eventsttl")
	if eventsttl == 0 {
		eventsttl = defaultEventsTTL
	}
	l, err := store.Store.Grant(store.Store.Ctx(), eventsttl)
	if err != nil {
		log.WithFields(log.Fields{
			"event.id":   ev.ID.String(),
			"event.name": ev.Name,
		}).WithError(err).Error("failed global broadcast, failed to get lease")
	}

	if _, err := store.Put(store.Store.Ctx(), k, string(v), clientv3.WithLease(l.ID)); err != nil {
		log.WithFields(log.Fields{
			"event.id":   ev.ID.String(),
			"event.name": ev.Name,
		}).WithError(err).Error("failed global broadcast, failed to write event to store")
	}
}

// globalListener listens for new global events in the store and rebroadcasts them locally
func globalListener(glStop chan struct{}) {
	defer glWg.Done()

	// Watch for new events being added to store
	wch := store.Store.Watch(store.Store.Ctx(), eventsPrefix, clientv3.WithPrefix(), clientv3.WithFilterDelete())
	for {
		select {
		case resp := <-wch:
			if resp.Canceled {
				return
			}
			for _, sev := range resp.Events {
				var ev api.Event
				if err := json.Unmarshal(sev.Kv.Value, &ev); err != nil {
					log.WithField("event.id", string(sev.Kv.Key)).WithError(err).Error("could not unmarshal global event")
					continue
				}
				if !uuid.Equal(ev.Origin, gdctx.MyUUID) {
					Broadcast(&ev)
				}
			}
		case <-glStop:
			return
		}
	}
}

// StartGlobal start the global events framework
// Should only be called after store is up.
func StartGlobal() error {
	ghID = Register(NewHandler(globalHandler))
	glStop = make(chan struct{}, 0)
	glWg.Add(1)
	go globalListener(glStop)

	return nil
}

// StopGlobal stops the global events framework
func StopGlobal() error {
	Unregister(ghID)
	close(glStop)
	glWg.Wait()

	return nil
}
