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
		Type:       api.BrickType(b.Type),
	}
}

func createSubvolInfo(sv *[]volume.Subvol) []api.Subvol {
	var subvols []api.Subvol

	for _, subvol := range *sv {
		var blist []api.BrickInfo
		for _, b := range subvol.Bricks {
			blist = append(blist, createBrickInfo(&b))
		}

		subvols = append(subvols, api.Subvol{
			Name:         subvol.Name,
			Type:         api.SubvolType(subvol.Type),
			Bricks:       blist,
			ReplicaCount: subvol.ReplicaCount,
			ArbiterCount: subvol.ArbiterCount,
		})
	}
	return subvols
}

func createVolumeInfoResp(v *volume.Volinfo) *api.VolumeInfo {

	return &api.VolumeInfo{
		ID:        v.ID,
		Name:      v.Name,
		Type:      api.VolType(v.Type),
		Transport: v.Transport,
		DistCount: v.DistCount,
		State:     api.VolState(v.State),
		Options:   v.Options,
		Subvols:   createSubvolInfo(&v.Subvols),
	}
}
