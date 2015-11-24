// Package volume contains some types associated with GlusterFS volumes that will be used in GlusterD
package volume

import (
	"bytes"
	"encoding/json"
	"path/filepath"

	"github.com/gluster/glusterd2/utils"
	"github.com/pborman/uuid"

	log "github.com/Sirupsen/logrus"
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
func NewVolumeEntry(req *VolCreateRequest) *Volinfo {
	v := NewVolinfo()

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

	v.Bricks = newBrickEntries(req.Bricks, v.ID, req.Force)
	if v.Bricks == nil {
		return nil
	}
	return v
}

// newBrickEntries returns list of initialized Brickinfo objects using list of
// bricks
func newBrickEntries(bricks []string, volId uuid.UUID, force bool) []Brickinfo {
	var b []Brickinfo
	var b1 Brickinfo
	for _, brick := range bricks {
		var e error
		hostname, path := utils.ParseHostAndBrickPath(brick)
		if len(hostname) == 0 || len(path) == 0 {
			return nil
		}
		//TODO : Check for peer hosts first, otherwise look for local
		//address
		local, err := utils.IsLocalAddress(hostname)
		if err != nil {
			log.WithField("Host", hostname).Error(err.Error())
			return nil
		}
		if local == false {
			log.WithField("Host", hostname).Error("Host is not local")
			return nil
		}

		b1.Hostname = hostname
		b1.Path, e = filepath.Abs(path)
		if e != nil {
			log.Error("Failed to convert the brickpath to absolute path")
			return nil
		}
		if utils.ValidateBrickPathLength(b1.Path) != 0 || utils.ValidateBrickSubDirLength(b1.Path) != 0 {
			return nil
		}
		if isBrickPathAvailable(b1.Hostname, b1.Path) != 0 {
			return nil
		}
		e = utils.ValidateBrickPathStats(b1.Path, b1.Hostname, force)
		if e != nil {
			//TODO: Need to communicate back the error
			return nil
		}
		e = utils.ValidateXattrSupport(b1.Path, b1.Hostname, volId, force)
		if e != nil {
			//TODO: Need to communicate back the error
			return nil
		}
		//TODO : Add validation to check whether file system support
		//extended attributes
		b = append(b, b1)
	}
	return b
}

// isBrickPathAvailable validates whether the brick is consumed by other
// volume
func isBrickPathAvailable(hostname string, brickPath string) int {
	volumes, e := GetVolumes()
	if e != nil || volumes == nil {
		// In case cluster doesn't have any volumes configured yet,
		// treat this as success
		log.Debug("Failed to retrieve volumes")
		return 0
	}
	for _, v := range volumes {
		for _, b := range v.Bricks {
			if b.Hostname == hostname && b.Path == brickPath {
				log.Error("Brick is already used by ", v.Name)
				return -1
			}
		}
	}
	return 0
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
