package transaction

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

// NeverStop can be used in UntilStop to make it never stop
var NeverStop <-chan struct{} = make(chan struct{})

// UntilStop loops until stop channel is closed, running f every d duration
func UntilStop(f func(), d time.Duration, stop <-chan struct{}) {
	var (
		t       *time.Timer
		timeout bool
	)

	for {
		select {
		case <-stop:
			return
		default:
		}
		func() {
			defer HandlePanic()
			f()
		}()
		t = ResetTimer(t, d, timeout)
		select {
		case <-stop:
			return
		case <-t.C:
			timeout = true
		}
	}
}

// ResetTimer avoids allocating a new timer if one is already in use
func ResetTimer(t *time.Timer, dur time.Duration, timeout bool) *time.Timer {
	if t == nil {
		return time.NewTimer(dur)
	}
	if !t.Stop() && !timeout {
		<-t.C
	}
	t.Reset(dur)
	return t
}

// HandlePanic simply recovers from a panic and logs an error.
func HandlePanic() {
	if r := recover(); r != nil {
		callers := getCallers()
		log.WithFields(log.Fields{
			"panic":   r,
			"callers": callers,
		}).Error("recovered from panic")
	}
}

func getCallers() (callers string) {
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return
		}
		callers += fmt.Sprintf("%v:%v\n", file, line)
	}
	return
}
