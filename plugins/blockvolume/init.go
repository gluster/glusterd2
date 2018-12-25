package blockvolume

import (
	// initialise all block providers
	_ "github.com/gluster/glusterd2/plugins/blockvolume/blockprovider/gluster-block"
	_ "github.com/gluster/glusterd2/plugins/blockvolume/blockprovider/gluster-loopback"
)
