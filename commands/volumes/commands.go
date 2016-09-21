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

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
	registerVolCreateStepFuncs()
	registerVolDeleteStepFuncs()
	registerVolStartStepFuncs()
	registerVolStopStepFuncs()
}
