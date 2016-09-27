package daemon

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/gluster/glusterd2/errors"

	log "github.com/Sirupsen/logrus"
)

type Daemon interface {
	Name() string       // Name of daemon
	Path() string       // absolute path to the binary
	Args() string       // args to pass to binary during start
	SocketFile() string // path to socket file to connect to
	PidFile() string    // path to pidfile
}

func Start(d Daemon, wait bool) error {
	cmd := exec.Command(d.Path(), d.Args())
	err := cmd.Start()
	if err != nil {
		return err
	}

	if wait == true {
		// Wait for the process to exit
		err = cmd.Wait()
		return err
	}

	// Check if pidfile exists
	pid, err := ReadPidFromFile(d.PidFile())
	if err == nil {
		// Check if process is running
		_, err := GetProcess(pid)
		if err == nil {
			return errors.ErrProcessAlreadyRunning
		}
	}

	log.WithFields(log.Fields{
		"name":       d.Name(),
		"args":       d.Args(),
		"pid":        cmd.Process.Pid,
		"pidfile":    d.PidFile(),
		"socketfile": d.SocketFile(),
	}).Info("Started daemon successfully")

	return nil
}

func Stop(d Daemon, force bool) error {

	pid, err := ReadPidFromFile(d.PidFile())
	if err != nil {
		return err
	}

	process, err := GetProcess(pid)
	if err != nil {
		return err
	}

	if force == true {
		err = process.Kill()
	} else {
		err = process.Signal(syscall.SIGTERM)
	}

	// TODO: Do this under some lock ?
	_ = os.Remove(d.PidFile())

	if err != nil {
		// TODO: log
	}

	return nil
}
