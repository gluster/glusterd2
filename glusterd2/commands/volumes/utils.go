package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func createBrickInfo(b *brick.Brickinfo) api.BrickInfo {
	return api.BrickInfo{
		ID:         b.ID,
		Path:       b.Path,
		VolumeID:   b.VolumeID,
		VolumeName: b.VolumeName,
		NodeID:     b.NodeID,
		Hostname:   b.Hostname,
	}
}

func createVolumeInfoResp(v *volume.Volinfo) *api.VolumeInfo {

	blist := make([]api.BrickInfo, len(v.Bricks))
	for i, b := range v.Bricks {
		blist[i] = createBrickInfo(&b)
	}

	return &api.VolumeInfo{
		ID:           v.ID,
		Name:         v.Name,
		Type:         api.VolType(v.Type),
		Transport:    v.Transport,
		DistCount:    v.DistCount,
		ReplicaCount: v.ReplicaCount,
		State:        api.VolState(v.State),
		Options:      v.Options,
		Bricks:       blist,
	}
}
