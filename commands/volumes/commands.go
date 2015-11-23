// Package volumecommands implements the volume management commands
package volumecommands

import (
	"github.com/gluster/glusterd2/rest"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() rest.Routes {
	return rest.Routes{
		rest.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes/",
			HandlerFunc: volumeCreateHandler},
		rest.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: volumeDeleteHandler},
		rest.Route{
			Name:        "VolumeInfo",
			Method:      "GET",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: volumeInfoHandler},
		rest.Route{
			Name:        "VolumeList",
			Method:      "GET",
			Pattern:     "/volumes/",
			HandlerFunc: volumeListHandler},
		rest.Route{
			Name:        "VolumeStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/start",
			HandlerFunc: volumeStartHandler},
		rest.Route{
			Name:        "VolumeStop",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/stop",
			HandlerFunc: volumeStopHandler},
	}
}
