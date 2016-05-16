// Package volumecommands implements the volume management commands
package volumecommands

import (
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Txns returns command transaction steps. Required for the Command interface.
func (c *Command) Txns() *transaction.Txns {
	return &transaction.Txns{
		transaction.RegisterTxn("VolumeCreate",
			validateVolumeCreate,
			generateVolfiles,
			storeVolume,
			rollBackVolumeCreate),
	}

}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		rest.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes",
			Version:     1,
			HandlerFunc: volumeCreateHandler},
		rest.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			Version:     1,
			HandlerFunc: volumeDeleteHandler},
		rest.Route{
			Name:        "VolumeInfo",
			Method:      "GET",
			Pattern:     "/volumes/{volname}",
			Version:     1,
			HandlerFunc: volumeInfoHandler},
		rest.Route{
			Name:        "VolumeList",
			Method:      "GET",
			Pattern:     "/volumes",
			Version:     1,
			HandlerFunc: volumeListHandler},
		rest.Route{
			Name:        "VolumeStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/start",
			Version:     1,
			HandlerFunc: volumeStartHandler},
		rest.Route{
			Name:        "VolumeStop",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/stop",
			Version:     1,
			HandlerFunc: volumeStopHandler},
	}
}
