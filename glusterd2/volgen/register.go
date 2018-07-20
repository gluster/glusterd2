package volgen

import (
	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

type clusterVolfileFunc func(*Volfile, []*volume.Volinfo, uuid.UUID)
type volumeVolfileFunc func(*Volfile, *volume.Volinfo, uuid.UUID)
type brickVolfileFunc func(*Volfile, *brick.Brickinfo, *volume.Volinfo, uuid.UUID)

type volfileFunc struct {
	name      string
	fn        interface{}
	nodeLevel bool
}

var (
	clusterVolfiles []volfileFunc
	volumeVolfiles  []volfileFunc
	brickVolfiles   []volfileFunc
)

func registerClusterVolfile(name string, cvf clusterVolfileFunc, nodeLevel bool) {
	clusterVolfiles = append(clusterVolfiles, volfileFunc{name, cvf, nodeLevel})
}

func registerVolumeVolfile(name string, vvf volumeVolfileFunc, nodeLevel bool) {
	volumeVolfiles = append(volumeVolfiles, volfileFunc{name, vvf, nodeLevel})
}

func registerBrickVolfile(name string, bvf brickVolfileFunc, nodeLevel bool) {
	brickVolfiles = append(brickVolfiles, volfileFunc{name, bvf, nodeLevel})
}
