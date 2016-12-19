package daemon

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/errors"

	log "github.com/Sirupsen/logrus"
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
	Args() string

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
}

// Start function starts the daemon located at path returned by Path() with
// args returned by Args() function. If the pidfile to the daemon exists, the
// contents are read to determine if the daemon is already running. If it
// is already running, errors.ErrProcessAlreadyRunning is returned.
// When wait == true, this function can be used to spawn short term processes
// which will be waited on for completion before this function returns.
func Start(d Daemon, wait bool) error {

	log.WithFields(log.Fields{
		"name": d.Name(),
		"path": d.Path(),
		"args": d.Args(),
	}).Debug("Starting daemon.")

	// Check if pidfile exists
	pid, err := ReadPidFromFile(d.PidFile())
	if err == nil {
		// Check if process is running
		_, err := GetProcess(pid)
		if err == nil {
			return errors.ErrProcessAlreadyRunning
		}
	}

	args := strings.Fields(d.Args())
	cmd := exec.Command(d.Path(), args...)
	err = cmd.Start()
	if err != nil {
		return err
	}

	if wait == true {
		// Wait for the child to exit
		errStatus := cmd.Wait()
		log.WithFields(log.Fields{
			"pid":    cmd.Process.Pid,
			"status": errStatus,
		}).Debug("Child exited")

		if errStatus != nil {
			// Child exited with error
			return errStatus
		}

		// It is assumed that the daemon will write it's pid to pidfile.
		// FIXME: When RPC infra is available, use that and make the
		// daemon tell glusterd2 that it's up and ready.
		pid, err = ReadPidFromFile(d.PidFile())
		if err != nil {
			log.WithFields(log.Fields{
				"pidfile": d.PidFile(),
				"error":   err.Error(),
			}).Error("Could not read pidfile")
			return err
		}

		log.WithFields(log.Fields{
			"name": d.Name(),
			"pid":  pid,
		}).Debug("Started daemon successfully")

		return nil
	}

	// If the process exits at some point later, do read it's
	// exit status. This should not let it be a zombie.
	go func() {
		err := cmd.Wait()
		log.WithFields(log.Fields{
			"name":   d.Name(),
			"pid":    cmd.Process.Pid,
			"status": err,
		}).Debug("Child exited.")
	}()

	return nil
}

// Stop function reads the PID from path returned by PidFile() and can
// terminate the process gracefully or forcefully.
// When force == false, a SIGTERM signal is sent to the daemon.
// When force == true, a SIGKILL signal is sent to the daemon.
func Stop(d Daemon, force bool) error {

	// It is assumed that the process d has written to pidfile
	pid, err := ReadPidFromFile(d.PidFile())
	if err != nil {
		return err
	}

	process, err := GetProcess(pid)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"name": d.Name(),
		"pid":  pid,
	}).Debug("Stopping daemon.")

	if force == true {
		err = process.Kill()
	} else {
		err = process.Signal(syscall.SIGTERM)
	}

	// TODO: Do this under some lock ?
	_ = os.Remove(d.PidFile())

	if err != nil {
		log.WithFields(log.Fields{
			"pid": pid,
		}).Error("Stopping daemon failed.")
	}

	return nil
}
