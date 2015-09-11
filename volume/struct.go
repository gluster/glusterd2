// Package volume contains some types associated with GlusterFS volumes that will be used in GlusterD
package volume

import (
	"bytes"
	"encoding/json"

	"code.google.com/p/go-uuid/uuid"
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
	ID   string
	Name string
	Type VolType

	Transport       string
	Bricks          []string
	DistCount       uint64
	ReplicaCount    uint16
	StripeCount     uint16
	DisperseCount   uint16
	RedundancyCount uint16

	Options map[string]string

	Status VolStatus

	Checksum uint64
	Version  uint64
}

// VolCreateRequest defines the parameters for creating a volume in the volume-create command
// TODO: This should probably be moved out of here.
type VolCreateRequest struct {
	Size            int      `json:"size,omitempty"` // omitempty till intelligent volume creation is handled
	Clusters        []string `json:"clusters,omitempty"`
	Name            string   `json:"name"`
	Transport       string   `json:"transport,omitempty"`
	DistCount       uint64   `json:"distcount,omitempty"`
	ReplicaCount    uint16   `json:"replica,omitempty"`
	StripeCount     uint16   `json:"stripecount,omitempty"`
	DisperseCount   uint16   `json:"dispersecount,omitempty"`
	RedundancyCount uint16   `json:"redundancycount,omitempty"`
	Snapshot        struct {
		Enable bool    `json:"enable"`
		Factor float32 `json:"factor"`
	} `json:"snapshot, omitempty"`
	Bricks []string `json:"bricks"`
}

// NewVolinfo returns an empty Volinfo
func NewVolinfo() *Volinfo {
	v := new(Volinfo)
	v.Options = make(map[string]string)

	return v
}

// NewVolumeEntry returns an initialized Volinfo using the given parameters
func NewVolumeEntry(req VolCreateRequest) *Volinfo {
	v := NewVolinfo()

	v.ID = uuid.NewUUID().String()
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

	v.Bricks = req.Bricks

	return v
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
