package store

import (
	"context"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	// LivenessKeyPrefix is the prefix in store where peers publish
	// their liveness information.
	LivenessKeyPrefix = "alive/"
)

// IsNodeAlive returns true and pid if the node specified is alive as seen by the store
func (s *GDStore) IsNodeAlive(peerID interface{}) (int, bool) {

	var keySuffix string

	switch peerID.(type) {
	case uuid.UUID:
		keySuffix = peerID.(uuid.UUID).String()
	case string:
		keySuffix = peerID.(string)
		if uuid.Parse(keySuffix) == nil {
			return 0, false
		}
	default:
		return 0, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := LivenessKeyPrefix + keySuffix
	resp, err := s.Get(ctx, key)
	if err != nil {
		return 0, false
	}

	if resp.Count == 1 {
		pid, err := strconv.Atoi(string(resp.Kvs[0].Value))
		if err != nil {
			log.WithError(err).Error("failed to parse pid")
			return 0, false
		}
		return pid, true
	}
	return 0, false
}

// GetAliveNodes will return a map of all alive nodes. It will contain
// peerID and corresponding process id. It uses single etcd query
// to get all alive nodes.
func (s *GDStore) GetAliveNodes(ctx context.Context) map[string]int {
	peerStatus := map[string]int{}

	resp, err := s.Get(ctx, LivenessKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return peerStatus
	}

	for _, kv := range resp.Kvs {
		peerID := path.Base(string(kv.Key))
		pid, err := strconv.Atoi(string(kv.Value))
		if err == nil {
			peerStatus[peerID] = pid
		}
	}

	return peerStatus
}

// AreNodesAlive returns true if all given nodes are alive.
func (s *GDStore) AreNodesAlive(ctx context.Context, peerIDs ...uuid.UUID) bool {
	var nodeIDs []string

	for _, peerID := range peerIDs {
		nodeIDs = append(nodeIDs, peerID.String())

	}

	peerStatus := s.GetAliveNodes(ctx)

	if len(nodeIDs) > len(peerStatus) {
		return false
	}

	for _, nodeID := range nodeIDs {
		if _, ok := peerStatus[nodeID]; !ok {
			return false
		}
	}

	return true
}

func (s *GDStore) publishLiveness() error {
	// publish liveness of this instance into the store
	key := LivenessKeyPrefix + gdctx.MyUUID.String()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.Put(ctx, key, strconv.Itoa(os.Getpid()), clientv3.WithLease(s.Session.Lease()))

	return err
}

func (s *GDStore) revokeLiveness() error {
	// revoke liveness (to be invoked during graceful shutdowns)
	key := LivenessKeyPrefix + gdctx.MyUUID.String()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.Delete(ctx, key)

	return err
}
