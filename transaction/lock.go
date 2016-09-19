package transaction

import (
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/store"

	"github.com/pborman/uuid"
)

const (
	lockPrefix = store.GlusterPrefix + "locks/"
)

// CreateLockStepFunc returns the registry IDs ofr StepFuncs which lock/unlock the given key
// If existing StepFuncs are not found, new funcs are created and registered.
func CreateLockStepFunc(key string) (string, string, error) {
	lockFuncID := key + ".Lock"
	unlockFuncID := key + ".Unlock"

	_, lockFuncFound := GetStepFunc(lockFuncID)
	_, unlockFuncFound := GetStepFunc(unlockFuncID)

	if lockFuncFound && unlockFuncFound {
		return lockFuncID, unlockFuncID, nil
	}

	key = lockPrefix + key
	locker, err := gdctx.Store.NewLock(key, nil)
	if err != nil {
		return "", "", err
	}

	if !lockFuncFound {
		lockFunc := func(c TxnCtx) error {
			log := c.Logger().WithField("key", key)

			_, err := locker.Lock(nil)
			if err != nil {
				log.WithError(err).Debug("failed to lock")
			} else {
				log.Debug("locked obtained")
			}

			return err
		}
		RegisterStepFunc(lockFunc, lockFuncID)
	}

	if !unlockFuncFound {
		unlockFunc := func(c TxnCtx) error {
			log := c.Logger().WithField("key", key)

			err := locker.Unlock()
			if err != nil {
				log.WithError(err).Error("failed to release lock")
			} else {
				log.Debug("lock released")
			}

			return err
		}
		RegisterStepFunc(unlockFunc, unlockFuncID)
	}

	return lockFuncID, unlockFuncID, nil
}

// CreateLockSteps retuns a lock and an unlock Step which lock/unlock the given key
func CreateLockSteps(key string) (*Step, *Step, error) {
	lockFunc, unlockFunc, err := CreateLockStepFunc(key)
	if err != nil {
		return nil, nil, err
	}

	lockStep := &Step{lockFunc, unlockFunc, []uuid.UUID{gdctx.MyUUID}}
	unlockStep := &Step{unlockFunc, "", []uuid.UUID{gdctx.MyUUID}}

	return lockStep, unlockStep, nil
}
