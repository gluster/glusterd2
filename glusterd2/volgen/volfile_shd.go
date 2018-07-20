package volgen

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateShdVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, peerid uuid.UUID) {
	volfile.FileName = "gluster/glustershd"
	shd := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("glustershd")

	for _, vol := range clusterinfo {
		if vol.State != volume.VolStarted {
			continue
		}

		filters := clusterGraphFilters{subvolTypes: []volume.SubvolType{
			volume.SubvolReplicate,
			volume.SubvolDisperse,
		}}
		clusterGraph(volfile, shd, vol, peerid, &filters)
	}
}

func init() {
	registerClusterVolfile("glustershd", generateShdVolfile, false)
}
