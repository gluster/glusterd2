// Package volume contains some types associated with GlusterFS volumes that will be used in GlusterD
package volume

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// VolStatus is the current status of a volume
type VolStatus uint16

const (
	// VolCreated should be set only for a volume that has been just created
	VolCreated VolStatus = iota
	// VolStarted should be set only for volumes that are running
	VolStarted
	// VolStopped should be set only for volumes that are not running, excluding newly created volumes
	VolStopped
)

var (
	ValidateBrickEntriesFunc   = ValidateBrickEntries
	validateBrickPathStatsFunc = utils.ValidateBrickPathStats
	NewVolinfoFunc             = NewVolinfo
	absFilePath                = filepath.Abs
	NewBrickEntriesFunc        = NewBrickEntries
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

	Status VolStatus

	Checksum uint64
	Version  uint64
	Bricks   []Brickinfo
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

// Brickinfo represents the information of a brick
type Brickinfo struct {
	Hostname string
	Path     string
	ID       uuid.UUID
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
func NewBrickEntries(bricks []string) ([]Brickinfo, error) {
	var brickInfos []Brickinfo
	var b Brickinfo
	var e error
	for _, brick := range bricks {
		hostname, path, err := utils.ParseHostAndBrickPath(brick)
		if err != nil {
			return nil, err
		}
		b.Hostname = hostname
		b.Path, e = absFilePath(path)
		if e != nil {
			log.Error("Failed to convert the brickpath to absolute path")
			return nil, e
		}
		b.ID, e = peer.GetPeerIDByAddrF(hostname)
		if e != nil {
			return nil, e
		}

		brickInfos = append(brickInfos, b)
	}
	return brickInfos, nil
}

// ValidateBrickEntries validates the brick list
func ValidateBrickEntries(bricks []Brickinfo, volID uuid.UUID, force bool) (int, error) {

	for _, brick := range bricks {
		if !uuid.Equal(brick.ID, gdctx.MyUUID) {
			continue
		}

		local, err := utils.IsLocalAddress(brick.Hostname)
		if err != nil {
			log.WithField("Host", brick.Hostname).Error(err.Error())
			return http.StatusInternalServerError, err
		}
		if local == false {
			log.WithField("Host", brick.Hostname).Error("Host is not local")
			return http.StatusBadRequest, errors.ErrBrickNotLocal
		}
		err = utils.ValidateBrickPathLength(brick.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = utils.ValidateBrickSubDirLength(brick.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = isBrickPathAvailable(brick.Hostname, brick.Path)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = validateBrickPathStatsFunc(brick.Path, brick.Hostname, force)
		if err != nil {
			return http.StatusBadRequest, err
		}
		err = utils.ValidateXattrSupport(brick.Path, brick.Hostname, volID, force)
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

	for _, b := range v.Bricks {
		nodes = append(nodes, b.ID)
	}
	return nodes
}
