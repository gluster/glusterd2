package transaction

import (
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/store"

	"github.com/coreos/etcd/clientv3/concurrency"
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
	locker := concurrency.NewLocker(gdctx.Store.Session, key)

	lockFunc := func(c TxnCtx) error {
		log := c.Logger().WithField("key", key)

		locker.Lock()
		log.Debug("locked obtained")

		return nil
	}
	RegisterStepFunc(lockFunc, lockFuncID)

	unlockFunc := func(c TxnCtx) error {
		log := c.Logger().WithField("key", key)

		locker.Unlock()
		log.Debug("lock released")

		return nil
	}
	RegisterStepFunc(unlockFunc, unlockFuncID)

	return lockFuncID, unlockFuncID, nil
}

// CreateLockSteps returns a lock and an unlock Step which lock/unlock the given key
func CreateLockSteps(key string) (*Step, *Step, error) {
	lockFunc, unlockFunc, err := CreateLockStepFunc(key)
	if err != nil {
		return nil, nil, err
	}

	lockStep := &Step{lockFunc, unlockFunc, []uuid.UUID{gdctx.MyUUID}}
	unlockStep := &Step{unlockFunc, "", []uuid.UUID{gdctx.MyUUID}}

	return lockStep, unlockStep, nil
}
