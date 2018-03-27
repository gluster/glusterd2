package volumecommands

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func getRedundancy(disperse uint) uint {
	var temp, l, mask uint
	temp = disperse
	l = 0
	for temp = temp >> 1; temp != 0; temp = temp >> 1 {
		l = l + 1
	}
	mask = ^(1 << l)
	if red := disperse & mask; red != 0 {
		return red
	}
	return 1
}

func checkDisperseParams(req *api.SubvolReq, s *volume.Subvol) error {
	count := len(req.Bricks)

	if req.DisperseData > 0 {
		if req.DisperseCount > 0 && req.DisperseRedundancy > 0 {
			if req.DisperseCount != req.DisperseData+req.DisperseRedundancy {
				return errors.New("Disperse count should be equal to sum of disperse-data and redundancy")
			}
		} else if req.DisperseRedundancy > 0 {
			req.DisperseCount = req.DisperseData + req.DisperseRedundancy
		} else if req.DisperseCount > 0 {
			req.DisperseRedundancy = req.DisperseCount - req.DisperseData
		} else {
			if count-req.DisperseData >= req.DisperseData {
				return errors.New("Need redundancy count along with disperse-data")
			}
			req.DisperseRedundancy = count - req.DisperseData
			req.DisperseCount = count
		}
	}

	if req.DisperseCount <= 0 {
		if count < 3 {
			return errors.New("Number of bricks must be greater than 2")
		}
		req.DisperseCount = count
	}

	if req.DisperseRedundancy <= 0 {
		req.DisperseRedundancy = int(getRedundancy(uint(req.DisperseCount)))
	}

	if req.DisperseCount != count {
		return errors.New("Disperse count and the number of bricks must be same for a pure disperse volume")
	}

	if 2*req.DisperseRedundancy >= req.DisperseCount {
		return errors.New("Invalid redundancy value")
	}

	s.DisperseCount = req.DisperseCount
	s.RedundancyCount = req.DisperseRedundancy

	return nil
}
