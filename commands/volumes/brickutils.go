package volumecommands

import (
	"os/exec"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/daemon"
	"github.com/gluster/glusterd2/utils"

	"github.com/pborman/uuid"
)

// BrickStartMaxRetries represents maximum no. of attempts that will be made
// to start brick processes in case of port clashes.
const BrickStartMaxRetries = 3

// Until https://review.gluster.org/#/c/16200/ gets into a release.
// And this is fully safe too as no other well-known errno exists after 132
const anotherEADDRINUSE = syscall.Errno(0x9E) // 158

func errorContainsErrno(err error, errno syscall.Errno) bool {
	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	status, ok := exiterr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	if status.ExitStatus() != int(errno) {
		return false
	}
	return true
}

// These functions are used in vol-create, vol-expand and vol-shrink (TBD)

func startBrick(b brick.Brickinfo) error {

	brickDaemon, err := brick.NewGlusterfsd(b)
	if err != nil {
		return err
	}

	for i := 0; i < BrickStartMaxRetries; i++ {
		err = daemon.Start(brickDaemon, true)
		if err != nil {
			if errorContainsErrno(err, syscall.EADDRINUSE) || errorContainsErrno(err, anotherEADDRINUSE) {
				// Retry iff brick failed to start because of port being in use.
				// Allow the previous instance to cleanup and exit
				time.Sleep(1 * time.Second)
			} else {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func stopBrick(b brick.Brickinfo) error {

	brickDaemon, err := brick.NewGlusterfsd(b)
	if err != nil {
		return err
	}

	err = daemon.Stop(brickDaemon, true)
	if err != nil {
		return err
	}

	return nil
}

func nodesFromBricks(bricks []string) ([]uuid.UUID, error) {

	var nodes []uuid.UUID
	var present bool
	for _, b := range bricks {
		present = false

		host, _, err := utils.ParseHostAndBrickPath(b)
		if err != nil {
			return nil, err
		}

		id := uuid.Parse(host)
		if id == nil {
			return nil, err
		}

		// if multiple bricks are on the same node, include the node
		// only once in the returned list
		for _, n := range nodes {
			if uuid.Equal(id, n) == true {
				present = true
				break
			}
		}

		if !present {
			nodes = append(nodes, id)
		}
	}

	return nodes, nil
}
