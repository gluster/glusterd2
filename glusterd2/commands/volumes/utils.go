package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func createSizeInfo(size *volume.SizeInfo) api.SizeInfo {
	return api.SizeInfo{
		Used:     size.Used,
		Free:     size.Free,
		Capacity: size.Capacity,
	}
}
