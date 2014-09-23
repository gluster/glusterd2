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

func NewVolinfo() *Volinfo {
	v := new(Volinfo)
	v.Options = make(map[string]string)

	return v
}

func New(volname, transport string, replica, stripe, disperse, redundancy uint16, bricks []string) *Volinfo {
	v := NewVolinfo()

	v.Id = uuid.NewUUID().String()
	v.Name = volname
	v.Transport = transport

	v.ReplicaCount = replica
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
