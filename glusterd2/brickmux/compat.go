package brickmux

import (
	"errors"
	"reflect"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/volume"
)

// ErrNoCompat is error returned when no compatible bricks can be found
var ErrNoCompat = errors.New("no compatible bricks found to be multiplexed onto")

// findCompatibleBrick first finds a compatible volume for multiplexing by
// comparing volumes having same set of volume options set and then picks a
// brick from the compatible volume.
func findCompatibleBrick(b *brick.Brickinfo, brickVolinfo *volume.Volinfo, volumes []*volume.Volinfo) (*brick.Brickinfo, error) {

	startedVolsPresent := false
	for _, v := range volumes {
		if v.State == volume.VolStarted {
			startedVolsPresent = true
			break
		}
	}

	var targetVolume *volume.Volinfo
	if !startedVolsPresent {
		// no started volumes present; allow bricks belonging to volume
		// that's about to be started to be multiplexed to the same
		// process
		targetVolume = brickVolinfo
	} else {
		for _, v := range volumes {
			if len(v.GetLocalBricks()) == 0 {
				// skip volumes that doesn't have bricks on this machine
				continue
			}
			if v.State != volume.VolStarted {
				// if volume isn't started, we can't multiplex.
				continue
			}
			if reflect.DeepEqual(v.Options, brickVolinfo.Options) {
				// compare volume options of volumes
				targetVolume = v
				break
			}
		}
	}

	if targetVolume == nil {
		return nil, ErrNoCompat
	}

	var targetBrick *brick.Brickinfo
	for _, localBrick := range targetVolume.GetLocalBricks() {
		if b.ID.String() == localBrick.ID.String() {
			continue
		}

		brickDaemon, _ := brick.NewGlusterfsd(localBrick)
		if running, _ := daemon.IsRunning(brickDaemon); running {
			targetBrick = &localBrick
			break
		}
	}

	if targetBrick == nil {
		return nil, ErrNoCompat
	}

	return targetBrick, nil
}
