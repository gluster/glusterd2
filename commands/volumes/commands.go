// Package volumecommands implements the volume management commands
package volumecommands

import (
	"github.com/gluster/glusterd2/servers/rest/route"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "VolumeCreate",
			Method:      "POST",
			Pattern:     "/volumes",
			Version:     1,
			HandlerFunc: volumeCreateHandler},
		route.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			Version:     1,
			HandlerFunc: volumeDeleteHandler},
		route.Route{
			Name:        "VolumeInfo",
			Method:      "GET",
			Pattern:     "/volumes/{volname}",
			Version:     1,
			HandlerFunc: volumeInfoHandler},
		route.Route{
			Name:        "VolumeStatus",
			Method:      "GET",
			Pattern:     "/volumes/{volname}/status",
			Version:     1,
			HandlerFunc: volumeStatusHandler},
		route.Route{
			Name:        "VolumeList",
			Method:      "GET",
			Pattern:     "/volumes",
			Version:     1,
			HandlerFunc: volumeListHandler},
		route.Route{
			Name:        "VolumeStart",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/start",
			Version:     1,
			HandlerFunc: volumeStartHandler},
		route.Route{
			Name:        "VolumeStop",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/stop",
			Version:     1,
			HandlerFunc: volumeStopHandler},
	}
}

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
}
