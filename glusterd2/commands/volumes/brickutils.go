package volumecommands

import (
	"errors"
	"fmt"
	"net/rpc"
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

// GetBrickRPCClient gets the rpc client corresponding to the brick
func GetBrickRPCClient(b *brick.Brickinfo) (*rpc.Client, error) {
	brickDaemon, err := brick.NewGlusterfsd(*b)
	if err != nil {
		return nil, err
	}

	client, err := daemon.GetRPCClient(brickDaemon)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// findCompatibleBrickInVol finds compatible brick for multiplexing from a specific volume
func findCompatibleBrickInVol(b *brick.Brickinfo, v *volume.Volinfo) (*brick.Brickinfo, error) {
	brickmuxLimit, err := cluster.MaxBricksPerGlusterfsd()
	if err != nil {
		log.WithError(err).Info("Couldn't get limit on brick multiplexing. Continue with no limits set on number of bricks per process.")
		brickmuxLimit = 0
	}

	for _, localBrick := range v.GetLocalBricks() {
		if b.Path == localBrick.Path {
			continue
		}

		if _, err := GetBrickRPCClient(&localBrick); err != nil {
			continue
		}

		port := pmap.RegistrySearch(localBrick.Path, pmap.GfPmapPortBrickserver)
		localBrickproc, err := brick.GetBrickProcessByPort(port)
		if err != nil {
			continue
		}

		if brickmuxLimit != 0 {
			if len(localBrickproc.Bricklist) >= brickmuxLimit {
				continue
			}
		}

		return &localBrick, nil
	}

	return nil, nil
}

// FindCompatibleBrick finds a compatible brick for multiplexing
func FindCompatibleBrick(b *brick.Brickinfo) (*brick.Brickinfo, error) {
	brickVolume, err := volume.GetVolume(b.VolumeName)
	if err != nil {
		return nil, err
	}

	compatBrick, err := findCompatibleBrickInVol(b, brickVolume)
	if compatBrick != nil {
		return compatBrick, nil
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
				compatBrick, _ := findCompatibleBrickInVol(b, vol)
				if compatBrick != nil {
					return compatBrick, nil
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
