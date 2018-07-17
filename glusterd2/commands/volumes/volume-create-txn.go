package volumecommands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
)

func undoStoreVolumeOnCreate(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volinfo").Debug("Failed to get key from store")
		return err
	}

	if err := deleteVolume(c); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Warn("Failed to delete volinfo from store")
	}

	return nil
}

func voltypeFromSubvols(req *api.VolCreateReq) volume.VolType {
	if len(req.Subvols) == 0 {
		return volume.Distribute
	}
	numSubvols := len(req.Subvols)

	// TODO: Don't know how to decide on Volume Type if each subvol is different
	// For now just picking the first subvols Type, which satisfies
	// most of today's needs
	switch req.Subvols[0].Type {
	case "replicate":
		if numSubvols > 1 {
			return volume.DistReplicate
		}
		return volume.Replicate
	case "disperse":
		if numSubvols > 1 {
			return volume.DistDisperse
		}
		return volume.Disperse
	case "distribute":
		return volume.Distribute
	default:
		return volume.Distribute
	}
}

func populateSubvols(volinfo *volume.Volinfo, req *api.VolCreateReq) error {
	var err error
	for idx, subvolreq := range req.Subvols {
		if subvolreq.ReplicaCount == 0 && subvolreq.Type == "replicate" {
			return errors.New("replica count not specified")
		}

		if subvolreq.ReplicaCount > 0 && (subvolreq.ReplicaCount+subvolreq.ArbiterCount) != len(subvolreq.Bricks) {
			return errors.New("invalid number of bricks specified. Number of bricks must be a multiple of replica (+arbiter) count")
		}

		name := fmt.Sprintf("%s-%s-%d", volinfo.Name, strings.ToLower(subvolreq.Type), idx)

		ty := volume.SubvolDistribute
		switch subvolreq.Type {
		case "replicate":
			ty = volume.SubvolReplicate
		case "disperse":
			ty = volume.SubvolDisperse
		default:
			ty = volume.SubvolDistribute
		}

		s := volume.Subvol{
			Name: name,
			ID:   uuid.NewRandom(),
			Type: ty,
		}

		if subvolreq.ArbiterCount != 0 {
			if subvolreq.ReplicaCount != 2 || subvolreq.ArbiterCount != 1 {
				return errors.New("for arbiter configuration, replica count must be 2 and arbiter count must be 1. The 3rd brick of the replica will be the arbiter")
			}
			s.ArbiterCount = 1
		}

		if subvolreq.ReplicaCount == 0 {
			s.ReplicaCount = 1
		} else {
			s.ReplicaCount = subvolreq.ReplicaCount
		}

		if subvolreq.DisperseCount != 0 || subvolreq.DisperseData != 0 || subvolreq.DisperseRedundancy != 0 {
			err = checkDisperseParams(&subvolreq, &s)
			if err != nil {
				return err
			}
		}
		s.Bricks, err = volume.NewBrickEntriesFunc(subvolreq.Bricks, volinfo.Name, volinfo.ID)
		if err != nil {
			return err
		}
		volinfo.Subvols = append(volinfo.Subvols, s)
	}

	return nil
}

func newVolinfo(req *api.VolCreateReq) (*volume.Volinfo, error) {

	volinfo := &volume.Volinfo{
		ID:        uuid.NewRandom(),
		Name:      req.Name,
		VolfileID: req.Name,
		State:     volume.VolCreated,
		Type:      voltypeFromSubvols(req),
		DistCount: len(req.Subvols),
		SnapList:  []string{},
		Auth: volume.VolAuth{
			Username: uuid.NewRandom().String(),
			Password: uuid.NewRandom().String(),
		},
	}

	if req.Options != nil {
		volinfo.Options = req.Options
	} else {
		volinfo.Options = make(map[string]string)
	}

	if req.Transport != "" {
		volinfo.Transport = req.Transport
	} else {
		volinfo.Transport = "tcp"
	}

	if err := populateSubvols(volinfo, req); err != nil {
		return nil, err
	}

	if req.Metadata != nil {
		volinfo.Metadata = req.Metadata
	} else {
		volinfo.Metadata = make(map[string]string)
	}

	return volinfo, nil
}

func createVolinfo(c transaction.TxnCtx) error {

	var req api.VolCreateReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	if volume.Exists(req.Name) {
		return gderrors.ErrVolExists
	}

	if len(req.Subvols) > 0 && req.Subvols[0].ArbiterCount > 0 {
		if req.Options == nil {
			req.Options = make(map[string]string)
		}
		req.Options["replicate.arbiter-count"] = fmt.Sprintf("%d", req.Subvols[0].ArbiterCount)
	}

	volinfo, err := newVolinfo(&req)
	if err != nil {
		return err
	}

	if err := validateXlatorOptions(req.Options, volinfo); err != nil {
		return err
	}

	if err := c.Set("volinfo", volinfo); err != nil {
		return err
	}

	if err := c.Set("bricks", volinfo.GetBricks()); err != nil {
		return err
	}

	allBricks, err := volume.GetAllBricksInCluster()
	if err != nil {
		return err
	}

	// Used by other peers to check if proposed bricks are already in use.
	// This check is however still prone to races. See issue #314
	if err := c.Set("all-bricks-in-cluster", allBricks); err != nil {
		return err
	}

	checks := brick.PrepareChecks(req.Force, req.Flags)
	err = c.Set("brick-checks", checks)

	return err
}
