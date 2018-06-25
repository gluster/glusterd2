package volgen2

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateTCPFuseVolfile(volfile *Volfile, vol *volume.Volinfo, peerid uuid.UUID) {
	volfile.FileName = vol.VolfileID

	dht := volfile.RootEntry.Add("debug/io-stats", vol, nil).SetName(vol.Name).
		Add("performance/io-threads", vol, nil).
		Add("performance/md-cache", vol, nil).
		Add("performance/open-behind", vol, nil).
		Add("performance/quick-read", vol, nil).
		Add("performance/io-cache", vol, nil).
		Add("performance/readdir-ahead", vol, nil).
		Add("performance/read-ahead", vol, nil).
		Add("performance/write-behind", vol, nil).
		Add("cluster/distribute", vol, nil)

	clusterGraph(volfile, dht, vol, peerid, nil)
}

func init() {
	registerVolumeVolfile("fuse", generateTCPFuseVolfile, false)
}
