package brickmux

import (
	"fmt"
	"os"

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
	var pidOnFile int
	log.WithField("brick", b.String()).Debug("get brick daemon for demultiplex process")
	brickDaemon, err := brick.NewGlusterfsd(b)
	if err != nil {
		return err
	}
	if pidOnFile, err = daemon.ReadPidFromFile(brickDaemon.PidFile()); err == nil {
		log.WithFields(log.Fields{"brick": b.String(),
			"pidfile": brickDaemon.PidFile()}).Error("Failed to read the pidfile")
		return err

	}
	brickDaemon.Socketfilepath, err = glusterdGetSockFromBrickPid(pidOnFile)
	if err != nil {
		log.WithFields(log.Fields{"brick": b.String(),
			"pid": pidOnFile}).Error("Failed to get the socket file of the glusterfsd process")
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
	log.WithField("brick", b.String()).Debug("detach request succeeded with result")

	// TODO: Find an alternative to substitute the sleep.
	// There might be some changes on glusterfsd side related to socket
	// files used while brick signout,
	// make appropriate changes once glusterfsd side is fixed.
	//time.Sleep(time.Millisecond * 200)

	log.WithField("brick", b.String()).Debug("deleting brick socket and pid file")
	os.Remove(brickDaemon.PidFile())

	return nil
}
