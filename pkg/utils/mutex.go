package utils

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const mutexLocked = 1 << iota

// MutexWithTry is a regular mutex with TryLock() functionality
type MutexWithTry struct {
	sync.Mutex
}

// TryLock is the non-blocking version of Lock(). It returns true if the lock
// is successfully acquired and false otherwise. This is functionally similar
// to Lock.acquire(false) from Python's threading module and tryLock() from
// Java.
func (mwt *MutexWithTry) TryLock() bool {
	// Refer golang source: src/sync/mutex.go
	//
	// type Mutex struct {
	//	state int32
	//	sema  uint32
	// }
	//
	// The first (unexported) member of sync.Mutex structure maintains the
	// internal state of the mutex. It's of type int32. The type, size and
	// offset of this state variable is very unlikely to change.
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(mwt)), 0, mutexLocked)
}
