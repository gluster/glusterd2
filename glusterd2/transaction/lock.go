package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pborman/uuid"
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

// createLockStepFunc returns the registry IDs of StepFuncs which lock/unlock the given key.
// If existing StepFuncs are not found, new funcs are created and registered.
func createLockStepFunc(key string) (string, string, error) {
	lockFuncID := key + ".Lock"
	unlockFuncID := key + ".Unlock"

	_, lockFuncFound := getStepFunc(lockFuncID)
	_, unlockFuncFound := getStepFunc(unlockFuncID)

	if lockFuncFound && unlockFuncFound {
		return lockFuncID, unlockFuncID, nil
	}

	key = lockPrefix + key
	locker := concurrency.NewMutex(store.Store.Session, key)

	lockFunc := func(c TxnCtx) error {

		ctx, cancel := context.WithTimeout(context.Background(), lockObtainTimeout)
		defer cancel()

		c.Logger().WithField("key", key).Debug("attempting to lock")
		err := locker.Lock(ctx)
		switch err {
		case nil:
			c.Logger().WithField("key", key).Debug("lock obtained")
		case context.DeadlineExceeded:
			// Propagate this all the way back to the client as a HTTP 409 response
			c.Logger().WithField("key", key).Debug("timeout: failed to obtain lock")
			err = ErrLockTimeout
		}

		return err
	}
	RegisterStepFunc(lockFunc, lockFuncID)

	unlockFunc := func(c TxnCtx) error {

		c.Logger().WithField("key", key).Debug("attempting to unlock")
		err := locker.Unlock(context.Background())
		if err == nil {
			c.Logger().WithField("key", key).Debug("lock unlocked")
		}

		return err
	}
	RegisterStepFunc(unlockFunc, unlockFuncID)

	return lockFuncID, unlockFuncID, nil
}

// CreateLockSteps returns a lock and an unlock Step which lock/unlock the given key
// TODO: Remove this function
func CreateLockSteps(key string) (*Step, *Step, error) {
	lockFunc, unlockFunc, err := createLockStepFunc(key)
	if err != nil {
		return nil, nil, err
	}

	lockStep := &Step{DoFunc: lockFunc, UndoFunc: unlockFunc, Nodes: []uuid.UUID{gdctx.MyUUID}, Skip: false}
	unlockStep := &Step{DoFunc: unlockFunc, UndoFunc: "", Nodes: []uuid.UUID{gdctx.MyUUID}, Skip: false}

	return lockStep, unlockStep, nil
}

// LockUnlockFunc is signature of functions used for distributed locking
// and unlocking.
type LockUnlockFunc func(ctx context.Context) error

// CreateLockFuncs creates and returns functions for distributed lock and
// unlock. This is similar to CreateLockSteps() but returns normal functions.
// TODO: Remove this function
func CreateLockFuncs(key string) (LockUnlockFunc, LockUnlockFunc) {

	key = lockPrefix + key
	locker := concurrency.NewMutex(store.Store.Session, key)

	// TODO: There is an opportunity for refactor here to re-use code
	// between CreateLockFunc and CreateLockSteps. This variant doesn't
	// have registry either.

	lockFunc := func(ctx context.Context) error {
		logger := gdctx.GetReqLogger(ctx)
		if logger == nil {
			logger = log.StandardLogger()
		}

		ctx, cancel := context.WithTimeout(ctx, lockObtainTimeout)
		defer cancel()

		logger.WithField("key", key).Debug("attempting to lock")
		err := locker.Lock(ctx)
		switch err {
		case nil:
			logger.WithField("key", key).Debug("lock obtained")
		case context.DeadlineExceeded:
			// Propagate this all the way back to the client as a HTTP 409 response
			logger.WithField("key", key).Debug("timeout: failed to obtain lock")
			err = ErrLockTimeout
		}

		return err
	}

	unlockFunc := func(ctx context.Context) error {
		logger := gdctx.GetReqLogger(ctx)
		if logger == nil {
			logger = log.StandardLogger()
		}

		logger.WithField("key", key).Debug("attempting to unlock")
		if err := locker.Unlock(context.Background()); err != nil {
			logger.WithField("key", key).WithError(err).Error("unlock failed")
			return err
		}

		logger.WithField("key", key).Debug("lock unlocked")
		return nil
	}

	return lockFunc, unlockFunc
}

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
	s, err := concurrency.NewSession(store.Store.Client, concurrency.WithTTL(lockTTL))
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
