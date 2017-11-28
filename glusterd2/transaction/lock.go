package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pborman/uuid"
)

const (
	lockPrefix        = "locks/"
	lockObtainTimeout = 5 * time.Second
)

// ErrLockTimeout is the error returned when lock could not be obtained
// and the request timed out
var ErrLockTimeout = errors.New("could not obtain lock: another conflicting transaction may be in progress")

// createLockStepFunc returns the registry IDs of StepFuncs which lock/unlock the given key.
// If existing StepFuncs are not found, new funcs are created and registered.
func createLockStepFunc(key string) (string, string, error) {
	lockFuncID := key + ".Lock"
	unlockFuncID := key + ".Unlock"

	_, lockFuncFound := GetStepFunc(lockFuncID)
	_, unlockFuncFound := GetStepFunc(unlockFuncID)

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
func CreateLockSteps(key string) (*Step, *Step, error) {
	lockFunc, unlockFunc, err := createLockStepFunc(key)
	if err != nil {
		return nil, nil, err
	}

	lockStep := &Step{lockFunc, unlockFunc, []uuid.UUID{gdctx.MyUUID}}
	unlockStep := &Step{unlockFunc, "", []uuid.UUID{gdctx.MyUUID}}

	return lockStep, unlockStep, nil
}
