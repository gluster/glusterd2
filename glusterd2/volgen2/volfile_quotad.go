package volgen2

import (
	"fmt"

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

		for i, subvol := range v.Subvols {
			if subvol.Type == volume.SubvolReplicate {
				name := fmt.Sprintf("%s-replicate-%d", v.Name, i)
				replicate := dht.Add("cluster/replicate", v, nil).SetName(name)
				for j, b := range subvol.Bricks {
					name := fmt.Sprintf("%s-replicate-%d-client-%d", v.Name, i, j)
					replicate.Add("protocol/client", v, &b).SetName(name)
				}
			}
		}
	}
}

func init() {
	registerClusterVolfile("quotad", generateQuotadVolfile, false)
}
