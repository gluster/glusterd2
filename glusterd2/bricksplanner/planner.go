package bricksplanner

import (
	"errors"
	"fmt"
	"path"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	config "github.com/spf13/viper"
)

func handleReplicaSubvolReq(req *api.VolCreateReq) error {
	if req.ReplicaCount < 2 {
		return nil
	}

	if req.ReplicaCount > 3 {
		return errors.New("invalid Replica Count")
	}
	req.SubvolType = "replicate"
	if req.ArbiterCount > 1 {
		return errors.New("invalid Arbiter Count")
	}

	return nil
}

func handleDisperseSubvolReq(req *api.VolCreateReq) error {
	if req.DisperseCount == 0 && req.DisperseDataCount == 0 && req.DisperseRedundancyCount == 0 {
		return nil
	}

	req.SubvolType = "disperse"

	if req.DisperseDataCount > 0 && req.DisperseRedundancyCount <= 0 {
		return errors.New("disperse redundancy count is required")
	}

	if req.DisperseDataCount > 0 {
		req.DisperseCount = req.DisperseDataCount + req.DisperseRedundancyCount
	}

	// Calculate Redundancy Value
	if req.DisperseRedundancyCount <= 0 {
		req.DisperseRedundancyCount = volume.GetRedundancy(uint(req.DisperseCount))
	}

	if req.DisperseDataCount <= 0 {
		req.DisperseDataCount = req.DisperseCount - req.DisperseRedundancyCount
	}

	if 2*req.DisperseRedundancyCount >= req.DisperseCount {
		return errors.New("invalid redundancy count")
	}

	return nil
}

// Based on the provided values like replica count, distribute count etc,
// brick layout will be created. Peer and device information for bricks are
// not available with the layout
func getBricksLayout(req *api.VolCreateReq) ([]api.SubvolReq, error) {
	var err error
	bricksMountRoot := path.Join(config.GetString("rundir"), "/bricks")

	// If Distribute count is zero then automatically decide
	// the distribute count based on available size in each device
	// TODO: Auto find the distribute count
	numSubvols := 1
	if req.DistributeCount > 0 {
		numSubvols = req.DistributeCount
	}

	// User input will be in MBs, convert to KBs for all
	// internal usage
	subvolSize := deviceutils.MbToKb(req.Size)
	if numSubvols > 1 {
		subvolSize = subvolSize / uint64(numSubvols)
	}

	if req.SnapshotReserveFactor < 1 {
		return nil, errors.New("invalid Snapshot Reserve Factor")
	}

	// Default Subvol Type
	req.SubvolType = "distribute"

	// Validations if replica and arbiter sub volume
	err = handleReplicaSubvolReq(req)
	if err != nil {
		return nil, err
	}

	// Validations if disperse sub volume
	err = handleDisperseSubvolReq(req)
	if err != nil {
		return nil, err
	}

	subvolplanner, exists := subvolPlanners[req.SubvolType]
	if !exists {
		return nil, errors.New("subvolume type not supported")
	}

	// Initialize the planner
	subvolplanner.Init(req, subvolSize)

	var subvols []api.SubvolReq

	// Create a Bricks layout based on replica count and
	// other details. Brick Path, PeerID information will
	// be added later.
	for i := 0; i < numSubvols; i++ {
		var bricks []api.BrickReq
		for j := 0; j < subvolplanner.BricksCount(); j++ {
			eachBrickSize := subvolplanner.BrickSize(j)
			brickType := subvolplanner.BrickType(j)
			eachBrickTpSize := uint64(float64(eachBrickSize) * req.SnapshotReserveFactor)

			bricks = append(bricks, api.BrickReq{
				Type:           brickType,
				Path:           fmt.Sprintf("%s/%s/subvol%d/brick%d/brick", bricksMountRoot, req.Name, i+1, j+1),
				Mountdir:       "/brick",
				TpName:         fmt.Sprintf("tp_%s_s%d_b%d", req.Name, i+1, j+1),
				LvName:         fmt.Sprintf("brick_%s_s%d_b%d", req.Name, i+1, j+1),
				Size:           eachBrickSize,
				TpSize:         eachBrickTpSize,
				TpMetadataSize: deviceutils.GetPoolMetadataSize(eachBrickTpSize),
				FsType:         "xfs",
				MntOpts:        "rw,inode64,noatime,nouuid",
			})
		}

		subvols = append(subvols, api.SubvolReq{
			Type:          req.SubvolType,
			Bricks:        bricks,
			ReplicaCount:  req.ReplicaCount,
			ArbiterCount:  req.ArbiterCount,
			DisperseCount: req.DisperseCount,
		})
	}

	return subvols, nil
}

// PlanBricks creates the brick layout with chosen device and size information
func PlanBricks(req *api.VolCreateReq) error {
	availableVgs, err := getAvailableVgs(req)
	if err != nil {
		return err
	}

	if len(availableVgs) == 0 {
		return errors.New("no devices registered or available for allocating bricks")
	}

	subvols, err := getBricksLayout(req)
	if err != nil {
		return err
	}

	zones := make(map[string]struct{})

	for idx, sv := range subvols {
		// If zones overlap is not specified then do not
		// reset the zones map so that other subvol bricks
		// will not get allocated in the same zones
		if req.SubvolZonesOverlap {
			zones = make(map[string]struct{})
		}

		// For the list of bricks, first try to utilize all the
		// unutilized devices, Once all the devices are used, then try
		// with device with expected space available.
		numBricksAllocated := 0
		for bidx, b := range sv.Bricks {
			totalsize := b.TpSize + b.TpMetadataSize

			for _, vg := range availableVgs {
				_, zoneUsed := zones[vg.Zone]
				if vg.AvailableSize >= totalsize && !zoneUsed && !vg.Used {
					subvols[idx].Bricks[bidx].PeerID = vg.PeerID
					subvols[idx].Bricks[bidx].VgName = vg.Name
					subvols[idx].Bricks[bidx].DevicePath = "/dev/" + vg.Name + "/" + b.LvName

					zones[vg.Zone] = struct{}{}
					numBricksAllocated++
					vg.AvailableSize -= totalsize
					vg.Used = true
					break
				}
			}
		}

		// All bricks allocation not satisfied since only fresh devices are
		// considered. Now consider all devices with available space
		if len(sv.Bricks) == numBricksAllocated {
			continue
		}

		// Try allocating for remaining bricks, No fresh device is available
		// but enough space is available in the devices
		for bidx := numBricksAllocated; bidx < len(sv.Bricks); bidx++ {
			b := sv.Bricks[bidx]
			totalsize := b.TpSize + b.TpMetadataSize

			for _, vg := range availableVgs {
				_, zoneUsed := zones[vg.Zone]
				if vg.AvailableSize >= totalsize && !zoneUsed {
					subvols[idx].Bricks[bidx].PeerID = vg.PeerID
					subvols[idx].Bricks[bidx].VgName = vg.Name
					subvols[idx].Bricks[bidx].DevicePath = "/dev/" + vg.Name + "/" + b.LvName

					zones[vg.Zone] = struct{}{}
					numBricksAllocated++
					vg.AvailableSize -= totalsize
					vg.Used = true
					break
				}
			}
		}

		// If the devices are not available as it is required for Volume.
		if len(sv.Bricks) != numBricksAllocated {
			return errors.New("no space available or all the devices are not registered")
		}
	}

	req.Subvols = subvols
	return nil
}
