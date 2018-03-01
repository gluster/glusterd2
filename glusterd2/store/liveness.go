package store

import (
	"context"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
)

const (
	// LivenessKeyPrefix is the prefix in store where peers publish
	// their liveness information.
	LivenessKeyPrefix = "alive/"
)

// IsNodeAlive returns true if the node specified is alive as seen by the store
func (s *GDStore) IsNodeAlive(nodeID interface{}) bool {

	var keySuffix string

	switch nodeID.(type) {
	case uuid.UUID:
		keySuffix = nodeID.(uuid.UUID).String()
	case string:
		keySuffix = nodeID.(string)
		if uuid.Parse(keySuffix) == nil {
			return false
		}
	default:
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := LivenessKeyPrefix + keySuffix
	resp, err := s.Get(ctx, key)
	if err != nil {
		return false
	}

	return resp.Count == 1
}

func (s *GDStore) publishLiveness() error {
	// publish liveness of this instance into the store
	key := LivenessKeyPrefix + gdctx.MyUUID.String()
	_, err := s.Put(context.TODO(), key, "", clientv3.WithLease(s.Session.Lease()))

	return err
}

func (s *GDStore) revokeLiveness() error {
	// revoke liveness (to be invoked during graceful shutdowns)
	key := LivenessKeyPrefix + gdctx.MyUUID.String()
	_, err := s.Delete(context.TODO(), key)

	return err
}
