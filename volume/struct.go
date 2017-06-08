// Package volume contains some types associated with GlusterFS volumes that will be used in GlusterD
package volume

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// VolState is the current status of a volume
type VolState uint16

const (
	// VolCreated should be set only for a volume that has been just created
	VolCreated VolState = iota
	// VolStarted should be set only for volumes that are running
	VolStarted
	// VolStopped should be set only for volumes that are not running, excluding newly created volumes
	VolStopped
)

var (
	// ValidateBrickEntriesFunc validates the brick list
	ValidateBrickEntriesFunc   = ValidateBrickEntries
	validateBrickPathStatsFunc = utils.ValidateBrickPathStats
	absFilePath                = filepath.Abs
	// NewBrickEntriesFunc creates the brick list
	NewBrickEntriesFunc = NewBrickEntries
)

// VolType is the status of the volume
type VolType uint16

const (
	// Distribute is a plain distribute volume
	Distribute VolType = iota
	// Replicate is plain replicate volume
	Replicate
	// Disperse is a plain erasure coded volume
	Disperse
	// DistReplicate is a distribute-replicate volume
	DistReplicate
	// DistDisperse is a distribute-'erasure coded' volume
	DistDisperse
)

// Volinfo repesents a volume
type Volinfo struct {
	ID           uuid.UUID
	Name         string
	Type         VolType
	Transport    string
	DistCount    int
	ReplicaCount int
	Options      map[string]string
	Status       VolState
	Checksum     uint64
	Version      uint64
	Bricks       []brick.Brickinfo
	Auth         VolAuth
}

// VolAuth represents username and password used by trusted/internal clients
type VolAuth struct {
	Username string
	Password string
}

// VolStatus represents collective status of the bricks that make up the volume
type VolStatus struct {
	Brickstatuses []brick.Brickstatus
	// TODO: Add further fields like memory usage, brick filesystem, fd consumed,
	// clients connected etc.
}

// NewBrickEntries creates the brick list
func NewBrickEntries(bricks []string, volName string, volID uuid.UUID) ([]brick.Brickinfo, error) {
	var brickInfos []brick.Brickinfo
	var binfo brick.Brickinfo

	for _, b := range bricks {
		host, path, e := utils.ParseHostAndBrickPath(b)
		if e != nil {
			return nil, e
		}

		binfo.Path, e = absFilePath(path)
		if e != nil {
			log.Error("Failed to convert the brickpath to absolute path")
			return nil, e
		}

		u := uuid.Parse(host)
		if u != nil {
			// Host specified is UUID
			binfo.NodeID = u
			p, e := peer.GetPeerF(host)
			if e != nil {
				return nil, e
			}
			binfo.Hostname = p.Addresses[0]
		} else {
			binfo.NodeID, e = peer.GetPeerIDByAddrF(host)
			if e != nil {
				return nil, e
			}
			binfo.Hostname = host
		}

		binfo.VolumeName = volName
		binfo.VolumeID = volID

		brickInfos = append(brickInfos, binfo)
	}
	return brickInfos, nil
}

// ValidateBrickEntries validates the brick list
func ValidateBrickEntries(bricks []brick.Brickinfo, volID uuid.UUID, force bool) (int, error) {

	for _, b := range bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		local, err := utils.IsLocalAddress(b.Hostname)
		if err != nil {
			log.WithField("Host", b.Hostname).Error(err.Error())
			return http.StatusInternalServerError, err
		}
		if local == false {
			log.WithField("Host", b.Hostname).Error("Host is not local")
			return http.StatusBadRequest, errors.ErrBrickNotLocal
		}
		err = utils.ValidateBrickPathLength(b.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = utils.ValidateBrickSubDirLength(b.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = isBrickPathAvailable(b.Hostname, b.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = validateBrickPathStatsFunc(b.Path, b.Hostname, force)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = utils.ValidateXattrSupport(b.Path, b.Hostname, volID, force)
		if err != nil {
			return http.StatusBadRequest, err
		}
	}
	return 0, nil
}

func (v *Volinfo) String() string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	return out.String()
}

// Nodes returns the a list of nodes on which this volume has bricks
func (v *Volinfo) Nodes() []uuid.UUID {
	var nodes []uuid.UUID

	// This shouldn't be very inefficient for small slices.
	var present bool
	for _, b := range v.Bricks {
		// Add node to the slice only if it isn't present already
		present = false
		for _, n := range nodes {
			if uuid.Equal(b.NodeID, n) == true {
				present = true
				break
			}
		}

		if present == false {
			nodes = append(nodes, b.NodeID)
		}
	}
	return nodes
}
