package transaction

import (
	"github.com/gluster/glusterd2/context"

	"github.com/docker/libkv/store"
)

// CreateLockStepFunc returns a lock and an unlock StepFunc which lock/unlock the given key
func CreateLockStepFunc(key string) (StepFunc, StepFunc, error) {
	locker, err := context.Store.NewLock(key, nil)
	if err != nil {
		return nil, nil, err
	}

	lockFunc = func(StepArg) StepRet {
		_, err := locker.Lock(nil)
		return err
	}
	unlockFunc = func(StepArg) StepRet {
		_, err := locker.Unlock(nil)
		return err
	}
	return lockFunc, unlockFunc, nil
}

// CreateLockSteps retuns a lock and an unlock Step which lock/unlock the given key
func CreateLockSteps(key string) (Step, Step, error) {
	lockFunc, unlockFunc, err := CreateLockStepFunc(key)
	if err != nil {
		return nil, nil, err
	}

	lockStep = Step{lockFunc, unlockFunc, []string{Leader}}
	unlockStep = Step{unlockFunc, nil, []string{Leader}}

	return lockStep, unlockStep, nil
}
