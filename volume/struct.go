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
	// NewVolinfoFunc returns an empty Volinfo
	NewVolinfoFunc = NewVolinfo
	absFilePath    = filepath.Abs
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
	// Stripe is a plain stripe volume
	Stripe
	// Disperse is a plain erasure coded volume
	Disperse
	// DistReplicate is a distribute-replicate volume
	DistReplicate
	// DistStripe is  a distribute-stripe volume
	DistStripe
	// DistDisperse is a distribute-'erasure coded' volume
	DistDisperse
	// DistRepStripe is a distribute-replicate-stripe volume
	DistRepStripe
	// DistDispStripe is distrbute-'erasure coded'-stripe volume
	DistDispStripe
)

// Volinfo repesents a volume
type Volinfo struct {
	ID   uuid.UUID
	Name string
	Type VolType

	Transport       string
	DistCount       uint64
	ReplicaCount    uint16
	StripeCount     uint16
	DisperseCount   uint16
	RedundancyCount uint16

	Options map[string]string

	Status VolState

	Checksum uint64
	Version  uint64
	Bricks   []brick.Brickinfo
}

// VolStatus represents collective status of the bricks that make up the volume
type VolStatus struct {
	Brickstatuses []brick.Brickstatus
	// TODO: Add further fields like memory usage, brick filesystem, fd consumed,
	// clients connected etc.
}

// VolCreateRequest defines the parameters for creating a volume in the volume-create command
// TODO: This should probably be moved out of here.
type VolCreateRequest struct {
	Name            string   `json:"name"`
	Transport       string   `json:"transport,omitempty"`
	DistCount       uint64   `json:"distcount,omitempty"`
	ReplicaCount    uint16   `json:"replica,omitempty"`
	StripeCount     uint16   `json:"stripecount,omitempty"`
	DisperseCount   uint16   `json:"dispersecount,omitempty"`
	RedundancyCount uint16   `json:"redundancycount,omitempty"`
	Bricks          []string `json:"bricks"`
	Force           bool     `json:"force,omitempty"`
}

// NewVolinfo returns an empty Volinfo
func NewVolinfo() *Volinfo {
	v := new(Volinfo)
	v.Options = make(map[string]string)

	return v
}

// NewVolumeEntry returns an initialized Volinfo using the given parameters
func NewVolumeEntry(req *VolCreateRequest) (*Volinfo, error) {
	v := NewVolinfoFunc()
	if v == nil {
		return nil, errors.ErrVolCreateFail
	}
	v.ID = uuid.NewRandom()
	v.Name = req.Name
	if len(req.Transport) > 0 {
		v.Transport = req.Transport
	} else {
		v.Transport = "tcp"
	}
	if req.ReplicaCount == 0 {
		v.ReplicaCount = 1
	} else {
		v.ReplicaCount = req.ReplicaCount
	}
	v.StripeCount = req.StripeCount
	v.DisperseCount = req.DisperseCount
	v.RedundancyCount = req.RedundancyCount
	//TODO : Generate internal username & password

	return v, nil
}

// NewBrickEntries creates the brick list
func NewBrickEntries(bricks []string) ([]brick.Brickinfo, error) {
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
			binfo.ID = u
			p, e := peer.GetPeerF(host)
			if e != nil {
				return nil, e
			}
			binfo.Hostname = p.Addresses[0]
		} else {
			binfo.ID, e = peer.GetPeerIDByAddrF(host)
			if e != nil {
				return nil, e
			}
			binfo.Hostname = host
		}

		brickInfos = append(brickInfos, binfo)
	}
	return brickInfos, nil
}

// ValidateBrickEntries validates the brick list
func ValidateBrickEntries(bricks []brick.Brickinfo, volID uuid.UUID, force bool) (int, error) {

	for _, b := range bricks {
		if !uuid.Equal(b.ID, gdctx.MyUUID) {
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
			if uuid.Equal(b.ID, n) == true {
				present = true
				break
			}
		}

		if present == false {
			nodes = append(nodes, b.ID)
		}
	}
	return nodes
}
