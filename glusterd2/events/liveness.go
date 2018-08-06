package events

import (
	"strings"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

const (
	eventPeerDisconnectedStore = "peer.disconnected.store"
	eventPeerConnectedStore    = "peer.connected.store"
)

type livenessWatcher struct {
	stopCh chan struct{}
	wg     sync.WaitGroup
	stop   sync.Once
}

var lWatcher *livenessWatcher

// Watch watches the store for nodes that go down or come up and it
// broadcasts this information locally.
func (l *livenessWatcher) Watch() {
	defer l.wg.Done()
	wch := store.Store.Watch(store.Store.Ctx(), store.LivenessKeyPrefix,
		clientv3.WithPrefix(), clientv3.WithKeysOnly())
	for {
		select {
		case resp := <-wch:
			if resp.Canceled {
				return
			}
			for _, sev := range resp.Events {

				peerID := strings.TrimPrefix(
					string(sev.Kv.Key),
					store.LivenessKeyPrefix)

				var evName string
				switch sev.Type {
				case clientv3.EventTypePut:
					evName = eventPeerConnectedStore
					log.WithField("id", peerID).Info("peer connected to store")
				case clientv3.EventTypeDelete:
					evName = eventPeerDisconnectedStore
					log.WithField("id", peerID).Info("peer disconnected from store")
				default:
					continue
				}

				data := map[string]string{
					"peer.id": peerID,
				}

				Broadcast(New(evName, data, false))
			}
		case <-l.stopCh:
			return
		}
	}
}

//Stop will stop the livenessWatcher if it is running and waits for
//it to exit.
func (l *livenessWatcher) Stop() error {
	l.stop.Do(func() {
		close(l.stopCh)
		l.wg.Wait()
	})
	return nil
}

func startLivenessWatcher() {
	lWatcher = &livenessWatcher{
		stopCh: make(chan struct{}),
	}
	lWatcher.wg.Add(1)
	go lWatcher.Watch()
}

func stopLivenessWatcher() {
	if err := lWatcher.Stop(); err != nil {
		log.WithError(err).Errorf("got error in stopping liveness watcher")
	}
}
