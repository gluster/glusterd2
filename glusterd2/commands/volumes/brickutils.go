package volumecommands

import (
	"fmt"
	"reflect"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/cluster"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"
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

		log.Infof("Got brick process for port %d", port)

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

// FindCompatibleBrickProcess finds a compatible brick process for multiplexing
func FindCompatibleBrickProcess(b *brick.Brickinfo) (*brick.Glusterfsd, error) {
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
