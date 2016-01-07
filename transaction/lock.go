package transaction

import (
	"github.com/gluster/glusterd2/context"
)

// CreateLockStepFunc returns a lock and an unlock StepFunc which lock/unlock the given key
func CreateLockStepFunc(key string) (StepFunc, StepFunc, error) {
	locker, err := context.Store.NewLock(key, nil)
	if err != nil {
		return nil, nil, err
	}

	lockFunc := func(c *context.Context, a StepArg) (StepRet, error) {
		log := c.Log.WithField("key", key)
		log.Debug("locking key in store")

		_, err := locker.Lock(nil)

		log.WithError(err).Debug("failed to lock key in store")
		return nil, err
	}

	unlockFunc := func(c *context.Context, a StepArg) (StepRet, error) {
		log := c.Log.WithField("key", key)
		log.Debug("unlocking key in store")

		err := locker.Unlock()

		log.WithError(err).Debug("failed to unlock key in store")
		return nil, err
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
