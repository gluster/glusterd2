package volgen2

import (
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

func generateQuotadVolfile(volfile *Volfile, clusterinfo []*volume.Volinfo, nodeid uuid.UUID) {
	volfile.FileName = "gluster/quotad"

	quotaOpts := make(map[string]string)

	for _, v := range clusterinfo {
		quotaOpts[v.Name+".volume-id"] = v.Name
	}

	quota := volfile.RootEntry.Add("features/quotad", nil, nil).SetName("quotad").SetExtraOptions(quotaOpts)

	for _, v := range clusterinfo {
		dht := quota.Add("cluster/distribute", v, nil).SetName(v.Name)
		clusterGraph(volfile, dht, v, nodeid, nil)
	}
}

func init() {
	registerClusterVolfile("quotad", generateQuotadVolfile, false)
}
