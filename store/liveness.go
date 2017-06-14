package store

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
)

const (
	ttlInterval        = 30
	ttlRefreshInterval = ttlInterval - 5
	livenessPrefix     = GlusterPrefix + "alive/"
)

var (
	livenessLeaseID     clientv3.LeaseID
	livenessStopRenewal chan struct{}
)

func publishLiveness(cli *clientv3.Client, livenessKey string) (clientv3.LeaseID, chan struct{}, error) {

	resp, err := cli.Grant(context.TODO(), ttlInterval)
	if err != nil {
		return 0, nil, err
	}

	key := livenessPrefix + livenessKey
	_, err = cli.Put(context.TODO(), key, "", clientv3.WithLease(resp.ID))
	if err != nil {
		return 0, nil, err
	}

	quitCh := make(chan struct{})
	go func(c *clientv3.Client, lid clientv3.LeaseID, revokeCh <-chan struct{}) {
		for {
			select {
			case <-quitCh:
				return
			case <-time.After(ttlRefreshInterval * time.Second):
				if resp, err := c.KeepAliveOnce(context.TODO(), lid); err != nil {
					log.WithError(err).WithField(
						"lease-id", lid).Error("failed to renew liveness lease")
				} else {
					log.WithFields(log.Fields{
						"lease": resp.ID,
						"ttl":   resp.TTL,
					}).Debug("liveness lease renewed successfully")
				}
			}
		}
	}(cli, resp.ID, quitCh)

	return resp.ID, quitCh, nil
}
