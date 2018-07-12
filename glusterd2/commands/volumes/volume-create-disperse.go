package volumecommands

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func checkDisperseParams(req *api.SubvolReq, s *volume.Subvol) error {
	count := len(req.Bricks)

	if req.DisperseData > 0 {
		if req.DisperseCount > 0 && req.DisperseRedundancy > 0 {
			if req.DisperseCount != req.DisperseData+req.DisperseRedundancy {
				return errors.New("disperse count should be equal to sum of disperse-data and redundancy")
			}
		} else if req.DisperseRedundancy > 0 {
			req.DisperseCount = req.DisperseData + req.DisperseRedundancy
		} else if req.DisperseCount > 0 {
			req.DisperseRedundancy = req.DisperseCount - req.DisperseData
		} else {
			if count-req.DisperseData >= req.DisperseData {
				return errors.New("need redundancy count along with disperse-data")
			}
			req.DisperseRedundancy = count - req.DisperseData
			req.DisperseCount = count
		}
	}

	if req.DisperseCount <= 0 {
		if count < 3 {
			return errors.New("number of bricks must be greater than 2")
		}
		req.DisperseCount = count
	}

	if req.DisperseRedundancy <= 0 {
		req.DisperseRedundancy = volume.GetRedundancy(uint(req.DisperseCount))
	}

	if req.DisperseCount != count {
		return errors.New("disperse count and the number of bricks must be same for a pure disperse volume")
	}

	if 2*req.DisperseRedundancy >= req.DisperseCount {
		return errors.New("invalid redundancy value")
	}

	s.DisperseCount = req.DisperseCount
	s.RedundancyCount = req.DisperseRedundancy

	return nil
}
