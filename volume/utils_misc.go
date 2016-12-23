package volume

import (
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/pborman/uuid"
)

var volCount uint64

func getRandVolume() *Volinfo {
	v := NewVolinfoFunc()

	v.ID = uuid.NewRandom()
	v.Name = fmt.Sprintf("volume-%d", atomic.AddUint64(&volCount, 1))
	v.Type = DistReplicate
	brickCount := rand.Intn(256) + 1
	for i := 1; i <= brickCount; i++ {
		//v.Bricks = append(v.Bricks, fmt.Sprintf("host:/brick-%d", i))
		v.Bricks[i].Hostname = "Host"
		v.Bricks[i].Path = fmt.Sprintf("/brick-%d", i)
		v.Bricks[i].ID = v.ID
	}
	v.DistCount = rand.Intn(256) + 1
	v.ReplicaCount = rand.Intn(10)
	v.StripeCount = rand.Intn(10)
	v.DisperseCount = rand.Intn(10)
	v.RedundancyCount = rand.Intn(10)

	v.Status = VolCreated

	v.Checksum = uint64(rand.Uint32())
	v.Version = 1

	return v
}
