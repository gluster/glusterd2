package volgen

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateRebalanceVolfile(volfile *Volfile, vol *volume.Volinfo, peerid uuid.UUID) {

	volfile.FileName = vol.Name + "/rebalance"

	dht := volfile.RootEntry.Add("debug/io-stats", vol, nil).SetName(vol.Name).
		Add("cluster/distribute", vol, nil)

	clusterGraph(volfile, dht, vol, peerid, nil)
}

func init() {
	registerVolumeVolfile("rebalance", generateRebalanceVolfile, false)
}
