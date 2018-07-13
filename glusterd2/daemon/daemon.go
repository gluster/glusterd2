package daemon

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/pkg/errors"

	log "github.com/sirupsen/logrus"
)

// Daemon interface should be implemented by individual daemons which wants
// glusterd to manage the lifecycle of the daemon.
type Daemon interface {

	// Name should return the name of the daemon. This will be primarily
	// used for logging.
	Name() string

	// Path should return absolute path to the binary/executable of
	// the daemon.
	Path() string

	// Args should return the arguments to be passed to the binary
	// during spawn.
	Args() []string

	// SocketFile should return the absolute path to the socket file which
	// will be used for inter process communication using Unix domain
	// socket. Please ensure that this function is deterministic i.e
	// it returns the same socket file path for same set of input
	// states.
	SocketFile() string

	// PidFile should return path to the pid file. This will be used by
	// the daemon framework to send signals (like SIGTERM/SIGKILL) to the
	// process. For now, it is the responsibility of the process to create
	// the pid file.
	PidFile() string

	// ID should return a unique identifier for the daemon.
	ID() string
}

// Start function starts the daemon located at path returned by Path() with
// args returned by Args() function. If the pidfile to the daemon exists, the
// contents are read to determine if the daemon is already running. If it
// is already running, errors.ErrProcessAlreadyRunning is returned.
// When wait == true, this function can be used to spawn short term processes
// which will be waited on for completion before this function returns.
func Start(d Daemon, wait bool, logger log.FieldLogger) error {

	logger.WithFields(log.Fields{
		"name": d.Name(),
		"path": d.Path(),
		"args": strings.Join(d.Args(), " "),
	}).Debug("Starting daemon.")
	events.Broadcast(newEvent(d, daemonStarting, 0))

	// Check if pidfile exists
	pid, err := ReadPidFromFile(d.PidFile())
	if err == nil {
		// Check if process is running
		_, err := GetProcess(pid)
		if err == nil {
			events.Broadcast(newEvent(d, daemonStarted, pid))
			return errors.ErrProcessAlreadyRunning
		}
	}

	cmd := exec.Command(d.Path(), d.Args()...)
	err = cmd.Start()
	if err != nil {
		events.Broadcast(newEvent(d, daemonStartFailed, 0))
		return err
	}

	if wait == true {
		// Wait for the child to exit
		errStatus := cmd.Wait()
		logger.WithFields(log.Fields{
			"pid":    cmd.Process.Pid,
			"status": errStatus,
		}).Debug("Child exited")

		if errStatus != nil {
			// Child exited with error
			events.Broadcast(newEvent(d, daemonStartFailed, 0))
			return errStatus
		}

		// It is assumed that the daemon will write it's pid to pidfile.
		// FIXME: When RPC infra is available, use that and make the
		// daemon tell glusterd2 that it's up and ready.
		pid, err = ReadPidFromFile(d.PidFile())
		if err != nil {
			logger.WithFields(log.Fields{
				"pidfile": d.PidFile(),
				"error":   err.Error(),
			}).Error("Could not read pidfile")
			events.Broadcast(newEvent(d, daemonStartFailed, 0))
			return err
		}

		logger.WithFields(log.Fields{
			"name": d.Name(),
			"pid":  pid,
		}).Debug("Started daemon successfully")
		events.Broadcast(newEvent(d, daemonStarted, pid))

	} else {
		// If the process exits at some point later, do read it's
		// exit status. This should not let it be a zombie.
		go func() {
			err := cmd.Wait()
			logger.WithFields(log.Fields{
				"name":   d.Name(),
				"pid":    cmd.Process.Pid,
				"status": err,
			}).Debug("Child exited.")
		}()
	}

	// Save daemon information in the store so it can be restarted
	if err := saveDaemon(d); err != nil {
		logger.WithField("name", d.Name()).WithError(err).Warn("failed to save daemon information into store, daemon may not be restarted on GlusterD restart")
	}

	return nil
}

// Kill function terminate the process gracefully or forcefully.
// When force == false, a SIGTERM signal is sent to the daemon.
// When force == true, a SIGKILL signal is sent to the daemon.
func Kill(pid int, force bool) error {

	process, err := GetProcess(pid)
	if err != nil {
		return err
	}

	if force {
		err = process.Kill()
	} else {
		err = process.Signal(syscall.SIGTERM)
	}

	return err
}

// Stop function reads the PID from path returned by PidFile() and can
// terminate the process gracefully or forcefully.
// When force == false, a SIGTERM signal is sent to the daemon.
// When force == true, a SIGKILL signal is sent to the daemon.
func Stop(d Daemon, force bool, logger log.FieldLogger) error {

	// It is assumed that the process d has written to pidfile
	pid, err := ReadPidFromFile(d.PidFile())
	if err != nil {
		return errors.ErrPidFileNotFound
	}

	logger.WithFields(log.Fields{
		"name": d.Name(),
		"pid":  pid,
	}).Debug("Stopping daemon.")
	events.Broadcast(newEvent(d, daemonStopping, pid))

	err = Kill(pid, force)

	// TODO: Do this under some lock ?
	_ = os.Remove(d.PidFile())

	if err != nil {
		logger.WithFields(log.Fields{
			"name": d.Name(),
			"pid":  pid,
		}).Error("Stopping daemon failed.")
		events.Broadcast(newEvent(d, daemonStopFailed, pid))
	} else {
		events.Broadcast(newEvent(d, daemonStopped, 0))
	}

	if err := DelDaemon(d); err != nil {
		logger.WithFields(log.Fields{
			"name": d.Name(),
			"pid":  pid,
		}).WithError(err).Warn("failed to delete daemon from store, it may be restarted on GlusterD restart")
	}

	return nil
}

// StartAllDaemons starts all previously running daemons when GlusterD restarts
func StartAllDaemons() {
	log.Debug("starting all daemons")
	events.Broadcast(events.New(daemonStartingAll, nil, false))

	ds, err := getDaemons()
	if err != nil {
		log.WithError(err).Warn("failed to get saved daemons, no daemons were started")
		events.Broadcast(events.New(daemonStartAllFailed, nil, false))
		return
	}

	for _, d := range ds {
		if err := Start(d, true, log.StandardLogger()); err != nil {
			log.WithField("name", d.Name()).WithError(err).Warn("failed to start daemon")
		}
	}
	events.Broadcast(events.New(daemonStartedAll, nil, false))
}

// Signal function reads the PID from path returned by PidFile() and
// sends the signal to that PID
func Signal(d Daemon, sig syscall.Signal, logger log.FieldLogger) error {

	// It is assumed that the process d has written to pidfile
	pid, err := ReadPidFromFile(d.PidFile())
	if err != nil {
		return err
	}

	process, err := GetProcess(pid)
	if err != nil {
		return err
	}

	logger.WithFields(log.Fields{
		"name":   d.Name(),
		"pid":    pid,
		"signal": sig,
	}).Debug("Signal to daemon.")

	err = process.Signal(sig)
	if err != nil {
		logger.WithFields(log.Fields{
			"name":   d.Name(),
			"pid":    pid,
			"signal": sig,
		}).Error("Sending signal to daemon failed.")
	}
	return nil
}
