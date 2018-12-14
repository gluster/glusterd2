package brickmux

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/glusterd2/volume"

	log "github.com/sirupsen/logrus"
)

func undoMultiplex(client *rpc.Client, b *brick.Brickinfo) {
	req := &brick.GfBrickOpReq{
		Name: b.Path,
		Op:   int(brick.OpBrickTerminate),
	}

	var rsp brick.GfBrickOpRsp
	client.Call("Brick.OpBrickTerminate", req, &rsp)
}

// Multiplex the specified brick onto a compatible running brick process.
func Multiplex(b brick.Brickinfo, v *volume.Volinfo, volumes []*volume.Volinfo, logger log.FieldLogger) error {

	targetBrick, err := findCompatibleBrick(&b, v, volumes)
	if err != nil {
		return err
	}

	// Kill already running fsd for the brick in consideration
	// Helpful in case of force start where no running glusterfsds are
	// effected and thus even after multiplexing, old glusterfsds
	// exist, take this action only when target brick for brick in
	// consideration is found or else current brick will go into offline
	// state.
	brickProc, err := brick.NewGlusterfsd(b)
	if err != nil {
		return err
	}
	pid, err := daemon.ReadPidFromFile(brickProc.PidFile())
	if err == nil {
		_, err = daemon.GetProcess(pid)
		if err == nil {
			err = b.TerminateBrick()
			if err != nil {
				return err
			}
		}
	}

	targetBrickProc, err := brick.NewGlusterfsd(*targetBrick)
	if err != nil {
		return err
	}

	targetBrickPort, err := pmap.RegistrySearch(targetBrick.Path)
	if err != nil {
		return err
	}

	// send attach request to target brick process

	client, err := daemon.GetRPCClient(targetBrickProc)
	if err != nil {
		return err
	}

	logger.WithFields(log.Fields{"brick": b.String(),
		"targetBrick": targetBrick.Path}).Info("found compatible brick process")
	logger.WithField("targetBrickSocketFile", targetBrickProc.SocketFile()).Debug("target brick socket file")

	volfileID := brick.GetVolfileID(b.VolumeName, b.Path)
	volfilePath, err := getBrickVolfilePath(volfileID)
	if err != nil {
		return err
	}

	req := &brick.GfBrickOpReq{
		Name: volfilePath,
		Op:   int(brick.OpBrickAttach),
	}

	logger.WithField("brick", b.Path).Debug("send brick attach RPC")
	var rsp brick.GfBrickOpRsp
	if err := client.Call("Brick.OpBrickAttach", req, &rsp); err != nil {
		return err
	}

	if rsp.OpRet != 0 {
		logger.WithError(err).WithField(
			"brick", b.String()).Error("failed to send attach RPC request")
		return fmt.Errorf("attach brick RPC request failed; OpRet = %d", rsp.OpRet)
	}
	logger.WithError(err).WithField(
		"brick", b.String()).Error("attach RPC request succeeded")

	brickProc, err = brick.NewGlusterfsd(b)
	if err != nil {
		undoMultiplex(client, &b)
		return err
	}

	// create Unix Domain Socket hardlink
	os.Remove(brickProc.SocketFile())
	if err := os.Link(targetBrickProc.SocketFile(), brickProc.SocketFile()); err != nil {
		undoMultiplex(client, &b)
		return err
	}

	// create duplicate pidfile for the multiplexed brick
	ok, pid := daemon.IsRunning(targetBrickProc)
	if !ok {
		return fmt.Errorf("brick process not running/found")
	}
	daemon.WritePidToFile(pid, brickProc.PidFile())

	// update pmap registry (this is redundant as each brick now signs in)
	pmap.RegistryExtend(b.Path, targetBrickPort, pid)

	return nil
}
