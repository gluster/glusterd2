package volgen

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateQuotadVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, peerid uuid.UUID) {
	volfile.FileName = "gluster/quotad"

	quotaOpts := make(map[string]string)

	for _, v := range clusterinfo {
		quotaOpts[v.Name+".volume-id"] = v.Name
	}

	quota := volfile.RootEntry.Add("features/quotad", nil, nil).SetName("quotad").SetExtraOptions(quotaOpts)

	for _, v := range clusterinfo {
		if v.State != volume.VolStarted {
			continue
		}

		//If quota is not enabled for volume, then skip those volumes
		val, exists := v.Options["quota.enable"]
		if exists && val == "on" {
			dht := quota.Add("cluster/distribute", v, nil).SetName(v.Name)
			clusterGraph(volfile, dht, v, peerid, nil)
		}
	}
}

func init() {
	registerClusterVolfile("quotad", generateQuotadVolfile, false)
}
