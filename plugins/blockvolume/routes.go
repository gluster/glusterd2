package blockvolume

import (
	"net/http"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/gluster/glusterd2/plugins/blockvolume/api"
	"github.com/gluster/glusterd2/plugins/blockvolume/hostvol"
)

// BlockVolume represents BlockVolume plugin
type BlockVolume struct {
	hostVolManager hostvol.HostingVolumeManager
	initOnce       sync.Once
}

// Name returns plugin name
func (b *BlockVolume) Name() string {
	b.Init()
	return "block-volume"
}

// RestRoutes returns list of REST API routes of BlockVolume to register with Glusterd.
func (b *BlockVolume) RestRoutes() route.Routes {
	b.Init()
	return route.Routes{
		{
			Name:         "BlockCreate",
			Method:       http.MethodPost,
			Pattern:      "/blockvolumes/{provider}",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.BlockVolumeCreateRequest)(nil)),
			ResponseType: utils.GetTypeString((*api.BlockVolumeCreateResp)(nil)),
			HandlerFunc:  b.CreateVolume,
		},
		{
			Name:        "BlockDelete",
			Method:      http.MethodDelete,
			Pattern:     "/blockvolumes/{provider}/{name}",
			Version:     1,
			HandlerFunc: b.DeleteVolume,
		},
		{
			Name:        "BlockList",
			Method:      http.MethodGet,
			Pattern:     "/blockvolumes/{provider}",
			Version:     1,
			HandlerFunc: b.ListBlockVolumes,
		},
		{
			Name:        "BlockGet",
			Method:      http.MethodGet,
			Pattern:     "/blockvolumes/{provider}/{name}",
			Version:     1,
			HandlerFunc: b.GetBlockVolume,
		},
	}
}

// RegisterStepFuncs registers all step functions
func (*BlockVolume) RegisterStepFuncs() {
	hostvol.RegisterBHVstepFunctions()
}

// Init will initialize the underlying HostVolume manager only once.
// calling it multiple times will do nothing
func (b *BlockVolume) Init() {
	b.initOnce.Do(func() {
		b.hostVolManager = hostvol.NewGlusterVolManager()
	})
}
