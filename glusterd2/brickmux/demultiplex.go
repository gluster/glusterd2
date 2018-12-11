package brickmux

import (
	"fmt"
	"os"
	"time"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"

	log "github.com/sirupsen/logrus"
)

// IsLastBrickInProc returns true if the brick specified is the last brick
// present in the brick process.
func IsLastBrickInProc(b brick.Brickinfo) bool {

	port, err := pmap.RegistrySearch(b.Path)
	if err != nil {
		return false
	}

	return len(pmap.GetBricksOnPort(port)) == 1
}

// Demultiplex sends a detach request to the brick process which the
// specified brick is multiplexed onto.
func Demultiplex(b brick.Brickinfo) error {

	log.WithField("brick", b.String()).Debug("get brick daemon for demultiplex process")
	brickDaemon, err := brick.NewGlusterfsd(b)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"brick": b.String(),
		"socketFile": brickDaemon.SocketFile()}).Debug("starting demultiplex process")

	client, err := daemon.GetRPCClient(brickDaemon)
	if err != nil {
		return err
	}

	req := &brick.GfBrickOpReq{
		Name: b.Path,
		Op:   int(brick.OpBrickTerminate),
	}

	log.WithField("brick", b.String()).Debug("detach request sent")
	var rsp brick.GfBrickOpRsp
	if err := client.Call("Brick.OpBrickTerminate", req, &rsp); err != nil {
		return err
	}

	if rsp.OpRet != 0 {
		return fmt.Errorf("detach brick RPC request failed; OpRet = %d", rsp.OpRet)
	}
	log.WithField("brick", b.String()).Debug("detach request succeded with result")

	// TODO: Find an alternative to substitute the sleep.
	// There might be some changes on glusterfsd side related to socket
	// files used while brick signout,
	// make appropriate changes once glusterfsd side is fixed.
	time.Sleep(time.Millisecond * 200)

	log.WithField("brick", b.String()).Debug("deleting brick socket and pid file")
	os.Remove(brickDaemon.PidFile())
	os.Remove(brickDaemon.SocketFile())
	log.WithField("brick", b.String()).Debug("deleted brick socket and pid file")

	return nil
}
