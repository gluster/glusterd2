package volgen2

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateTCPGfProxyFuseVolfile(volfile *Volfile, vol *volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gfproxy-client/" + vol.Name

	volfile.RootEntry.Add("debug/io-stats", vol, nil).SetName(vol.Name).
		Add("performance/write-behind", vol, nil).
		Add("protocol/client", vol, nil).SetExtraData(map[string]string{"brick.path": "gfproxyd-" + vol.Name, "brick.hostname": ""})
}

func generateGfproxydVolfile(volfile *Volfile, vol *volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gfproxyd/" + vol.Name

	dht := volfile.RootEntry.Add("protocol/server", vol, nil).SetExtraData(map[string]string{"brick.path": "", "brick.hostname": ""}).
		Add("debug/io-stats", vol, nil).SetName(vol.Name).
		Add("performance/io-threads", vol, nil).
		Add("performance/md-cache", vol, nil).
		Add("performance/open-behind", vol, nil).
		Add("performance/quick-read", vol, nil).
		Add("performance/io-cache", vol, nil).
		Add("performance/readdir-ahead", vol, nil).
		Add("performance/read-ahead", vol, nil).
		Add("cluster/distribute", vol, nil)

	clusterGraph(volfile, dht, vol, nodeid, nil)
}

func init() {
	registerVolumeVolfile("gfproxyd", generateGfproxydVolfile, false)
	registerVolumeVolfile("gfproxy-client", generateTCPGfProxyFuseVolfile, false)
}
