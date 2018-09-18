package pmap

import (
	"fmt"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// This change synchronizes port allocation on a node/machine when multiple
// glusterd2 instances are running (e2e tests). This solution is to be
// treated as stopgap fix. A long term and more robust solution is to let
// the bricks pick and bind on their own.

func (r *pmapRegistry) TryPortLock(port int) bool {

	if _, ok := r.portLockFds[port]; ok {
		// we already have a lock
		return false
	}

	portLockFile := fmt.Sprintf("/var/run/glusterd2_pmap_port_%d.lock", port)
	fd, err := syscall.Open(portLockFile,
		syscall.O_CREAT|syscall.O_WRONLY|syscall.O_CLOEXEC, 0666)
	if err != nil {
		return false
	}

	err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
	switch err {
	case nil:
		// keep the fd open if we get the lock, close otherwise
		r.portLockFds[port] = fd
		log.WithField("lock-file", portLockFile).Debug(
			"Obtained lock on pmap port lock file.")
		return true
	case syscall.EWOULDBLOCK:
		log.WithField("lock-file", portLockFile).Debug(
			"Failed to obtain lock on pmap port lock file.")
	}

	syscall.Close(fd)
	return false
}

func (r *pmapRegistry) PortUnlock(port int) {
	fd, ok := r.portLockFds[port]
	if !ok {
		return
	}
	syscall.Flock(fd, syscall.LOCK_UN)
	syscall.Close(fd)
	delete(r.portLockFds, port)
	if r.ports[port].State == PortFree || r.ports[port].State == PortForeign {
		portLockFile := fmt.Sprintf("/var/run/glusterd2_pmap_port_%d.lock", port)
		syscall.Unlink(portLockFile)
	}
}
