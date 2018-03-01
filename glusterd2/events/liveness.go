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

func startLivenessWatcher() {
	lWatcher = &livenessWatcher{
		stopCh: make(chan struct{}, 0),
	}
	lWatcher.wg.Add(1)
	go lWatcher.Watch()
}

func stopLivenessWatcher() {
	close(lWatcher.stopCh)
	lWatcher.wg.Wait()
}
