package volgen2

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateShdVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gluster/glustershd"
	shd := volfile.RootEntry.Add("debug/io-stats", nil, nil).SetName("glustershd")

	for _, vol := range clusterinfo {
		filters := clusterGraphFilters{subvolTypes: []volume.SubvolType{
			volume.SubvolReplicate,
			volume.SubvolDisperse,
		}}
		clusterGraph(volfile, shd, vol, nodeid, &filters)
	}
}

func init() {
	registerClusterVolfile("glustershd", generateShdVolfile, false)
}
