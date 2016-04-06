package transaction

import (
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/store"
)

const (
	lockPrefix = store.GlusterPrefix + "locks/"
)

// CreateLockStepFunc returns a lock and an unlock StepFunc which lock/unlock the given key
func CreateLockStepFunc(key string) (StepFunc, StepFunc, error) {
	key = lockPrefix + key
	locker, err := context.Store.NewLock(key, nil)
	if err != nil {
		return nil, nil, err
	}

	lockFunc := func(c *context.Context) error {
		log := c.Log.WithField("key", key)

		_, err := locker.Lock(nil)
		if err != nil {
			log.WithError(err).Debug("failed to lock")
		} else {
			log.Debug("locked obtained")
		}

		return err
	}

	unlockFunc := func(c *context.Context) error {
		log := c.Log.WithField("key", key)

		err := locker.Unlock()
		if err != nil {
			log.WithError(err).Error("failed to release lock")
		} else {
			log.Debug("lock released")
		}

		return err
	}
	return lockFunc, unlockFunc, nil
}

// CreateLockSteps retuns a lock and an unlock Step which lock/unlock the given key
func CreateLockSteps(key string) (*Step, *Step, error) {
	lockFunc, unlockFunc, err := CreateLockStepFunc(key)
	if err != nil {
		return nil, nil, err
	}

	lockStep := &Step{lockFunc, unlockFunc, []string{Leader}}
	unlockStep := &Step{unlockFunc, nil, []string{Leader}}

	return lockStep, unlockStep, nil
}
