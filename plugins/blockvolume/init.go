package blockvolume

import (
	// initialise all block providers
	_ "github.com/gluster/glusterd2/plugins/blockvolume/blockprovider/gluster-block"

	config "github.com/spf13/viper"
)

func init() {
	config.SetDefault("block-provider", "gluster-block")
}
