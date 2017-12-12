package daemon

import (
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/events"
)

type daemonEvent string

const (
	daemonStarting       daemonEvent = "daemon.starting"
	daemonStarted                    = "daemon.started"
	daemonStartFailed                = "daemon.startfailed"
	daemonStopping                   = "daemon.stopping"
	daemonStopped                    = "daemon.stopped"
	daemonStopFailed                 = "daemon.stopfailed"
	daemonStartingAll                = "daemon.startingall"
	daemonStartedAll                 = "daemon.startedall"
	daemonStartAllFailed             = "daemon.startallfailed"
)

// newEvent returns an event of given type with daemon data filled
func newEvent(d Daemon, e daemonEvent, pid int) *events.Event {
	data := map[string]string{
		"name":   d.Name(),
		"id":     d.ID(),
		"binary": d.Path(),
		"args":   d.Args(),
		"pid":    strconv.Itoa(pid),
	}

	return events.New(string(e), data, false)
}
