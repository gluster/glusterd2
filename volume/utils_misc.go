package volume

import (
	"fmt"
	"math/rand"
	"sync/atomic"

	"code.google.com/p/go-uuid/uuid"
)

var volCount uint64 = 0

func getRandVolume() *Volinfo {
	v := NewVolinfo()

	v.Id = uuid.NewUUID().String()
	v.Name = fmt.Sprintf("volume-%d", atomic.AddUint64(&volCount, 1))
	v.Type = DistReplicate
	brickCount := uint64(rand.Intn(256) + 1)
	for i := uint64(1); i <= brickCount; i += 1 {
		v.Bricks = append(v.Bricks, fmt.Sprintf("host:/brick-%d", i))
	}
	v.DistCount = uint64(rand.Intn(256) + 1)
	v.ReplicaCount = uint16(rand.Intn(10))
	v.StripeCount = uint16(rand.Intn(10))
	v.DisperseCount = uint16(rand.Intn(10))
	v.RedundancyCount = uint16(rand.Intn(10))

	v.Status = VolCreated

	v.Checksum = uint64(rand.Uint32())
	v.Version = 1

	return v
}
