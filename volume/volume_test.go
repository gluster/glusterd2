package volume

import (
	"testing"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/tests"
)

func TestNewVolumeEntry(t *testing.T) {
	v := NewVolinfo()

	tests.Assert(t, v.Options != nil)
	tests.Assert(t, len(v.ID) == 0)
}

func TestNewVolumeEntryFromEmptyRequest(t *testing.T) {
	req := new(VolCreateRequest)
	v := NewVolumeEntry(req)
	tests.Assert(t, v == nil)
}

func find(haystack []string, needle string) bool {

	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}

	return false
}

func TestNewVolumeEntryFromRequestBricks(t *testing.T) {
	context.Init()
	bricks := []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	brickPaths := []string{"/tmp/b1", "/tmp/b2"}
	hosts := []string{"127.0.0.1"}

	b := newBrickEntries(bricks, false)
	tests.Assert(t, b != nil)
	for _, brick := range b {
		tests.Assert(t, find(brickPaths, brick.Path))
		tests.Assert(t, find(hosts, brick.Hostname))
	}

}

func TestNewVolumeEntryFromRequest(t *testing.T) {
	context.Init()
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	v := NewVolumeEntry(req)
	tests.Assert(t, v.Name == "vol1")
	tests.Assert(t, v.Transport == "tcp")
	tests.Assert(t, v.ReplicaCount == 1)
	tests.Assert(t, len(v.ID) != 0)
	tests.Assert(t, len(v.Bricks) != 0)

}

func TestNewVolumeEntryFromRequestReplica(t *testing.T) {
	context.Init()
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	req.ReplicaCount = 3

	v := NewVolumeEntry(req)
	tests.Assert(t, v.ReplicaCount == 3)
}

func TestNewVolumeEntryFromRequestTransport(t *testing.T) {
	context.Init()
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Transport = "rdma"
	req.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}

	v := NewVolumeEntry(req)
	tests.Assert(t, v.Transport == "rdma")
}

func TestNewVolumeEntryFromRequestStripe(t *testing.T) {
	context.Init()
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	req.StripeCount = 2

	v := NewVolumeEntry(req)
	tests.Assert(t, v.StripeCount == 2)
}

func TestNewVolumeEntryFromRequestDisperse(t *testing.T) {
	context.Init()
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	req.DisperseCount = 2

	v := NewVolumeEntry(req)
	tests.Assert(t, v.DisperseCount == 2)
}

func TestNewVolumeEntryFromRequestRedundancy(t *testing.T) {
	context.Init()
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	req.RedundancyCount = 2

	v := NewVolumeEntry(req)
	tests.Assert(t, v.RedundancyCount == 2)
}
