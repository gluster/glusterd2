package brickmux

import (
	"errors"
	"reflect"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// ErrNoCompat is error returned when no compatible bricks can be found
var ErrNoCompat = errors.New("no compatible bricks found to be multiplexed onto")

// validateBmuxTarget validates whether target volume has any valid brick in
// which current brick can be multiplexed into. Useful in case when
// max-brick-per-process(> 0) has been set by user
func validateBmuxTarget(b *brick.Brickinfo, volinfo *volume.Volinfo, maxBricksPerProcess int) *brick.Brickinfo {
	var targetBrick *brick.Brickinfo
	for _, localbrick := range volinfo.GetLocalBricks() {
		if uuid.Equal(b.ID, localbrick.ID) {
			continue
		}

		brickDaemon, err := brick.NewGlusterfsd(localbrick)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume": volinfo.Name,
				"brick":  b.Path}).Error("failed to retrieve brick daemon object")
			continue
		}
		if running, _ := daemon.IsRunning(brickDaemon); !running {
			continue
		}

		// Look for port for the brick path of target volume
		port, err := pmap.RegistrySearch(localbrick.Path)
		if err != nil {
			continue
		}
		// return number of bricks already attached to a port
		numOfBricksBmuxedOnPort, err := pmap.GetNumOfBricksOnPort(port)
		if err != nil {
			continue
		}
		if numOfBricksBmuxedOnPort == maxBricksPerProcess {
			continue
		}
		targetBrick = &localbrick
		break
	}
	return targetBrick
}

// findCompatibleBrick first finds a compatible volume for multiplexing by
// comparing volumes having same set of volume options set and then picks a
// brick from the compatible volume.
func findCompatibleBrick(b *brick.Brickinfo, brickVolinfo *volume.Volinfo, volumes []*volume.Volinfo, maxBricksPerProcess int) (*brick.Brickinfo, error) {

	startedVolsPresent := false
	for _, v := range volumes {
		if v.State == volume.VolStarted {
			startedVolsPresent = true
			break
		}
	}

	var targetBrick *brick.Brickinfo
	var targetVolume *volume.Volinfo
	if !startedVolsPresent {
		// no started volumes present; allow bricks belonging to volume
		// that's about to be started to be multiplexed to the same
		// process
		targetVolume = brickVolinfo
		if maxBricksPerProcess > 0 {
			targetBrick = validateBmuxTarget(b, brickVolinfo, maxBricksPerProcess)
			if targetBrick == nil {
				return nil, ErrNoCompat
			}
		}
	} else {
		// atleast one started volume present
		for _, v := range volumes {
			localBricks := v.GetLocalBricks()
			if len(localBricks) == 0 {
				// skip volumes that doesn't have bricks on this machine
				continue
			}
			if v.State != volume.VolStarted {
				// if volume isn't started, we can't multiplex.
				continue
			}
			// compare volume options of volumes
			if reflect.DeepEqual(v.Options, brickVolinfo.Options) {
				targetVolume = v
				if maxBricksPerProcess > 0 {
					targetBrick = validateBmuxTarget(b, targetVolume, maxBricksPerProcess)
					if targetBrick == nil {
						continue
					}
				}
				break
			}
		}
		// If any of the started volumes doesn't qualify as a
		//targetBrick and a targetVolume,
		// then look for an eligible bricks in current volume.
		// Current volume is not present in started volumes list at this
		// step therefore need to evaluate target brick for current
		// volume seperately.
		if targetBrick == nil && maxBricksPerProcess > 0 {
			targetBrick = validateBmuxTarget(b, brickVolinfo, maxBricksPerProcess)
			if targetBrick == nil {
				return nil, ErrNoCompat
			}
		}
	}

	// Useful in case with no target Volume and no max-brick-per-process
	// constraint
	// consider multiplexing all bricks of current volume into first brick
	// of current volume. Handled seperately since current volume still not
	// considered in started volumes list. Executed when
	// max-bricks-per-process = 0.
	if targetVolume == nil {
		targetVolume = brickVolinfo
	}
	// if target volume exists and target brick has not been evaluated yet.
	// Useful in case of max-bricks-per-process  = 0.
	if targetBrick == nil {
		targetBrick = validateBmuxTarget(b, targetVolume, maxBricksPerProcess)
		if targetBrick == nil {
			return nil, ErrNoCompat
		}

	}
	return targetBrick, nil
}
