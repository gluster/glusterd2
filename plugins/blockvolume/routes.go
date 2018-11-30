package blockvolume

import (
	"net/http"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/gluster/glusterd2/plugins/blockvolume/api"
	"github.com/gluster/glusterd2/plugins/blockvolume/blockprovider"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

// BlockVolume represents BlockVolume plugin
type BlockVolume struct {
	blockProvider blockprovider.Provider
	initOnce      sync.Once
}

// Name returns underlying block provider name
func (b *BlockVolume) Name() string {
	b.mustInitBlockProvider()
	return b.blockProvider.ProviderName()
}

// RestRoutes returns list of REST API routes of BlockVolume to register with Glusterd.
func (b *BlockVolume) RestRoutes() route.Routes {
	b.mustInitBlockProvider()
	return route.Routes{
		{
			Name:         "BlockCreate",
			Method:       http.MethodPost,
			Pattern:      "/blockvolumes",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.BlockVolumeCreateRequest)(nil)),
			ResponseType: utils.GetTypeString((*api.BlockVolumeCreateResp)(nil)),
			HandlerFunc:  b.CreateVolume,
		},
		{
			Name:        "BlockDelete",
			Method:      http.MethodDelete,
			Pattern:     "/blockvolumes/{name}",
			Version:     1,
			HandlerFunc: b.DeleteVolume,
		},
		{
			Name:        "BlockList",
			Method:      http.MethodGet,
			Pattern:     "/blockvolumes",
			Version:     1,
			HandlerFunc: b.ListBlockVolumes,
		},
		{
			Name:        "BlockGet",
			Method:      http.MethodGet,
			Pattern:     "/blockvolumes/{name}",
			Version:     1,
			HandlerFunc: b.GetBlockVolume,
		},
	}
}

// RegisterStepFuncs registers all step functions
// Here it is a no-op func
func (*BlockVolume) RegisterStepFuncs() {

}

// mustInitBlockProvider will initialize the underlying block provider only once.
// calling it multiple times will do nothing
func (b *BlockVolume) mustInitBlockProvider() {
	b.initOnce.Do(func() {
		providerName := config.GetString("block-provider")
		provider, err := blockprovider.GetBlockProvider(providerName)
		if err != nil {
			log.WithError(err).Panic("failed in initializing block-volume provider")
		}
		b.blockProvider = provider
	})
}
