package volumecommands

import (
	"fmt"
	"reflect"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/cluster"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"

	log "github.com/sirupsen/logrus"
)

// findCompatibleBrickInVol finds compatible brick for multiplexing from a specific volume
func findCompatibleBrickProcInVol(b *brick.Brickinfo, v *volume.Volinfo) (*brick.Glusterfsd, error) {
	brickmuxLimit, err := cluster.MaxBricksPerGlusterfsd()
	if err != nil {
		log.WithError(err).Info("Couldn't get limit on brick multiplexing. Continue with no limits set on number of bricks per process.")
		brickmuxLimit = 0
	}

	for _, localBrick := range v.GetLocalBricks() {
		if b.Path == localBrick.Path {
			continue
		}

		port := pmap.RegistrySearch(localBrick.Path, pmap.GfPmapPortBrickserver)
		if port == 0 {
			// Couldn't find brick entry in portmap
			continue
		}

		localBrickProc, err := brick.GetBrickProcessByPort(port)
		if err != nil {
			continue
		}

		if brickmuxLimit != 0 {
			if len(localBrickProc.Bricklist) >= brickmuxLimit {
				continue
			}
		}

		_, err = daemon.GetRPCClient(localBrickProc)
		if err != nil {
			continue
		}

		return localBrickProc, nil
	}

	return nil, nil
}

// findCompatibleBrickProcess finds a compatible brick process for multiplexing
func findCompatibleBrickProcess(b *brick.Brickinfo) (*brick.Glusterfsd, error) {
	brickVolume, err := volume.GetVolume(b.VolumeName)
	if err != nil {
		return nil, err
	}

	compatBrickProc, err := findCompatibleBrickProcInVol(b, brickVolume)
	if compatBrickProc != nil {
		return compatBrickProc, nil
	}

	vols, err := volume.GetVolumes()
	if err != nil {
		return nil, err
	}

	for _, vol := range vols {
		if vol.Name == b.VolumeName {
			continue
		} else {
			if reflect.DeepEqual(vol.Options, brickVolume.Options) {
				compatBrickProc, _ := findCompatibleBrickProcInVol(b, vol)
				if compatBrickProc != nil {
					return compatBrickProc, nil
				}
			}
		}
	}
	return nil, nil
}

// RestartBricksInVolume is called upon node reboot or gd2 restart
func RestartBricksInVolume(v *volume.Volinfo) error {
	for _, b := range v.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(b)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("Restarting brick")

		pid, err := daemon.ReadPidFromFile(brickDaemon.PidFile())
		if err == nil {
			// Get list of brick processes to find the brick process for brick 'b'
			bps, err := brick.GetBrickProcesses()
			if err != nil {
				return err
			}

			var port int
			for _, bp := range bps {
				if bp.Pid != pid {
					continue
				}

				port = bp.Port
				break
			}
			// If the brick process is already running populate portmap and return
			if _, err := daemon.GetProcess(pid); err == nil {
				log.Infof("Adding brick %s to port %d", b.Path, port)
				pmap.RegistryExtend(port, b.Path, pmap.GfPmapPortBrickserver)
				continue
			}

			if err := daemon.DelDaemon(brickDaemon); err != nil {
				log.WithFields(log.Fields{
					"name": brickDaemon.Name(),
					"id":   brickDaemon.ID(),
				}).WithError(err).Warn("failed to delete stale brick entry from store")

				return err
			}

			log.Infof("Deleted brick daemon from store")

			if err := brick.DeleteBrickProcess(brickDaemon); err != nil {
				log.WithFields(log.Fields{
					"name": brickDaemon.Name(),
					"id":   brickDaemon.ID(),
				}).WithError(err).Warn("failed to delete stale brick process instance from store")

				return err
			}
		}

		if err := StartBrick(b); err != nil {
			return err
		}

	}

	return nil
}

// StartBrick starts a brick
func StartBrick(b brick.Brickinfo) error {

	log.WithFields(log.Fields{
		"volume": b.VolumeName,
		"brick":  b.String(),
	}).Info("Starting brick")

	brickmux, err := cluster.IsBrickMuxEnabled()
	if err != nil {
		return err
	}

	if !brickmux {
		if err := b.StartBrickProcess(); err != nil {
			return err
		}
		return nil
	}

	compatBrickProc, err := findCompatibleBrickProcess(&b)
	if err != nil {
		return err
	}

	if compatBrickProc != nil {
		log.Infof("Found compatible brick process with pid %d", compatBrickProc.Pid)

		client, err := daemon.GetRPCClient(compatBrickProc)
		if err != nil {
			return err
		}

		req := &brick.GfBrickOpReq{
			Name: volgen.GetBrickVolFileID(b.VolumeName, b.PeerID.String(), b.Path),
			Op:   int(brick.OpBrickAttach),
		}

		var rsp brick.GfBrickOpRsp
		err = client.Call("Brick.OpBrickAttach", req, &rsp)
		if err != nil || rsp.OpRet != 0 {
			log.WithError(err).WithField(
				"brick", b.String()).Error("failed to send attach RPC, starting brick process")
			if err := b.StartBrickProcess(); err != nil {
				return err
			}
		}

		pmap.RegistryExtend(compatBrickProc.Port, b.Path, pmap.GfPmapPortBrickserver)

		daemon.WritePidToFile(compatBrickProc.Pid, brick.GetPidFilePathForBrick(b))

		// Update brick process info in store
		compatBrickProc.Bricklist = compatBrickProc.AddBrick(b)

		if err := brick.UpdateBrickProcess(compatBrickProc); err != nil {
			log.WithField("name", compatBrickProc.Name()).WithError(err).Warn(
				"failed to save daemon information into store, daemon may not be restarted correctly on GlusterD restart")
			return err
		}
	} else {
		if err := b.StartBrickProcess(); err != nil {
			return err
		}
	}

	return nil
}

func nodesFromVolumeCreateReq(req *api.VolCreateReq) ([]uuid.UUID, error) {
	var nodesMap = make(map[string]bool)
	var nodes []uuid.UUID
	for _, subvol := range req.Subvols {
		for _, brick := range subvol.Bricks {
			if _, ok := nodesMap[brick.PeerID]; !ok {
				nodesMap[brick.PeerID] = true
				u := uuid.Parse(brick.PeerID)
				if u == nil {
					return nil, fmt.Errorf("Failed to parse peer ID: %s", brick.PeerID)
				}
				nodes = append(nodes, u)
			}
		}
	}
	return nodes, nil
}

func nodesFromVolumeExpandReq(req *api.VolExpandReq) ([]uuid.UUID, error) {
	var nodesMap = make(map[string]bool)
	var nodes []uuid.UUID
	for _, brick := range req.Bricks {
		if _, ok := nodesMap[brick.PeerID]; !ok {
			nodesMap[brick.PeerID] = true
			u := uuid.Parse(brick.PeerID)
			if u == nil {
				return nil, fmt.Errorf("Failed to parse peer ID: %s", brick.PeerID)
			}
			nodes = append(nodes, u)
		}
	}
	return nodes, nil
}
