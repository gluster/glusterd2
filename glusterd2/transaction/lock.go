package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3/concurrency"
	log "github.com/sirupsen/logrus"
)

const (
	lockPrefix        = "locks/"
	lockObtainTimeout = 5 * time.Second
	lockTTL           = 10
)

var (
	// ErrLockTimeout is the error returned when lock could not be obtained
	// and the request timed out
	ErrLockTimeout = errors.New("could not obtain lock: another conflicting transaction may be in progress")
	// ErrLockExists is returned when a lock already exists within the transaction
	ErrLockExists = errors.New("existing lock found for given lock ID")
)

// Locks are the collection of cluster wide transaction lock
type Locks map[string]*concurrency.Mutex

func (l Locks) lock(lockID string) error {
	var logger = log.WithField("lockID", lockID)

	// Ensure that no prior lock exists for the given lockID in this transaction
	if _, ok := l[lockID]; ok {
		return ErrLockExists
	}

	logger.Debug("attempting to obtain lock")

	key := lockPrefix + lockID
	s, err := concurrency.NewSession(store.Store.NamespaceClient, concurrency.WithTTL(lockTTL))
	if err != nil {
		return err
	}

	locker := concurrency.NewMutex(s, key)

	ctx, cancel := context.WithTimeout(store.Store.Ctx(), lockObtainTimeout)
	defer cancel()

	err = locker.Lock(ctx)
	switch err {
	case nil:
		logger.Debug("lock obtained")
		// Attach lock to the transaction
		l[lockID] = locker

	case context.DeadlineExceeded:
		logger.Debug("timeout: failed to obtain lock")
		// Propagate this all the way back to the client as a HTTP 409 response
		err = ErrLockTimeout

	default:
		logger.WithError(err).Error("failed to obtain lock")
	}

	return err
}

// Lock obtains a cluster wide transaction lock on the given lockID/lockIDs,
// and attaches the obtained locks to the transaction
func (l Locks) Lock(lockID string, lockIDs ...string) error {
	if err := l.lock(lockID); err != nil {
		return err
	}
	for _, id := range lockIDs {
		if err := l.lock(id); err != nil {
			return err
		}
	}
	return nil
}

// UnLock releases all cluster wide obtained locks
func (l Locks) UnLock(ctx context.Context) {
	for lockID, locker := range l {
		if err := locker.Unlock(ctx); err == nil {
			delete(l, lockID)
		}
	}
}
