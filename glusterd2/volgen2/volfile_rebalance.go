package volgen2

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateRebalanceVolfile(volfile *Volfile, vol *volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "rebalance/" + vol.Name

	dht := volfile.RootEntry.Add("debug/io-stats", vol, nil).SetName(vol.Name).
		Add("cluster/distribute", vol, nil)

	clusterGraph(volfile, dht, vol, nodeid, nil)
}

func init() {
	registerVolumeVolfile("rebalance", generateRebalanceVolfile, false)
}
