// Package volume contains some types associated with GlusterFS volumes that will be used in GlusterD
package volume

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"path/filepath"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
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
//go:generate stringer -type=VolType
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
	Auth         VolAuth // TODO: should not be returned to client
	GraphMap     map[string]string
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

// StringMap returns a map[string]string representation of Volinfo
func (v *Volinfo) StringMap() map[string]string {
	m := make(map[string]string)

	m["volume.id"] = v.ID.String()
	m["volume.name"] = v.Name
	m["volume.type"] = v.Type.String()
	m["volume.transport"] = v.Transport
	m["volume.auth.username"] = v.Auth.Username
	m["volume.auth.password"] = v.Auth.Password

	return m
}

// NewBrickEntries creates the brick list
func NewBrickEntries(bricks []string, volName string, volID uuid.UUID) ([]brick.Brickinfo, error) {
	var brickInfos []brick.Brickinfo
	var binfo brick.Brickinfo

	for _, b := range bricks {
		node, path, e := utils.ParseHostAndBrickPath(b)
		if e != nil {
			return nil, e
		}

		u := uuid.Parse(node)
		if u == nil {
			return nil, errors.New("Invalid UUID specified as host for brick")
		}

		p, e := peer.GetPeerF(node)
		if e != nil {
			return nil, e
		}

		binfo.NodeID = u
		// TODO: Have a better way to select peer address here
		binfo.Hostname, _, _ = net.SplitHostPort(p.Addresses[0])

		binfo.Path, e = absFilePath(path)
		if e != nil {
			log.Error("Failed to convert the brickpath to absolute path")
			return nil, e
		}

		binfo.VolumeName = volName
		binfo.VolumeID = volID
		binfo.ID = uuid.NewRandom()

		brickInfos = append(brickInfos, binfo)
	}
	return brickInfos, nil
}

// ValidateBrickEntries validates the brick list
func ValidateBrickEntries(bricks []brick.Brickinfo, volID uuid.UUID, force bool) (int, error) {

	var err error
	for _, b := range bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		err = utils.ValidateBrickPathLength(b.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = utils.ValidateBrickSubDirLength(b.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = isBrickPathAvailable(b.NodeID, b.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = validateBrickPathStatsFunc(b.Path, force)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = utils.ValidateXattrSupport(b.Path, volID, force)
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
