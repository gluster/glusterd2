package volume

import (
	"bytes"
	"encoding/json"

	"code.google.com/p/go-uuid/uuid"
)

type VolStatus uint16

const (
	VolCreated VolStatus = iota
	VolStarted
	VolStopped
)

type VolType uint16

const (
	Distribute VolType = iota
	Replicate
	Stripe
	Disperse
	DistReplicate
	DistStripe
	DistDisperse
	DistRepStripe
	DistDispStripe
)

type Volinfo struct {
	Id   string
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

type VolumeCreateRequest struct {
	Name            string `json:"name"`
	Transport       string `json:"transport,omitempty"`
	DistCount       uint64 `json:"distcount,omitempty"`
	ReplicaCount    uint16 `json:"replica,omitempty"`
	StripeCount     uint16 `json:"stripecount,omitempty"`
	DisperseCount   uint16 `json:"dispersecount,omitempty"`
	RedundancyCount uint16 `json:"redundancycount,omitempty"`

	Bricks []string `json:"bricks"`
}

func NewVolinfo() *Volinfo {
	v := new(Volinfo)
	v.Options = make(map[string]string)

	return v
}

func New(volname, transport string, replica, stripe, disperse, redundancy uint16, bricks []string) *Volinfo {
	v := NewVolinfo()

	v.Id = uuid.NewUUID().String()
	v.Name = volname
	if len(transport) > 0 {
		v.Transport = transport
	} else {
		v.Transport = "tcp"
	}
	if replica == 0 {
		v.ReplicaCount = 1
	} else {
		v.ReplicaCount = replica
	}
	v.StripeCount = stripe
	v.DisperseCount = disperse
	v.RedundancyCount = redundancy

	v.Bricks = bricks

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
