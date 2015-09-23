package volume

import (
	"testing"

	"github.com/kshlm/glusterd2/tests"
)

func TestNewVolumeEntry(t *testing.T) {
	v := NewVolinfo()

	tests.Assert(t, v.Options != nil)
	tests.Assert(t, len(v.ID) == 0)
}

func TestNewVolumeEntryFromEmptyRequest(t *testing.T) {

	req := VolCreateRequest{}

	v := NewVolumeEntry(req)
	tests.Assert(t, len(v.Name) == 0)
}

func TestNewVolumeEntryFromRequest(t *testing.T) {

	req := VolCreateRequest{}
	req.Name = "vol1"

	v := NewVolumeEntry(req)
	tests.Assert(t, v.Name == "vol1")
	tests.Assert(t, v.Transport == "tcp")
	tests.Assert(t, v.ReplicaCount == 1)
	tests.Assert(t, len(v.ID) != 0)
	tests.Assert(t, len(v.Bricks) == 0)

}

func TestNewVolumeEntryFromRequestReplica(t *testing.T) {

	req := VolCreateRequest{}
	req.Name = "vol1"
	req.ReplicaCount = 3

	v := NewVolumeEntry(req)
	tests.Assert(t, v.ReplicaCount == 3)
}

func TestNewVolumeEntryFromRequestTransport(t *testing.T) {

	req := VolCreateRequest{}
	req.Transport = "rdma"

	v := NewVolumeEntry(req)
	tests.Assert(t, v.Transport == "rdma")
}

func TestNewVolumeEntryFromRequestStripe(t *testing.T) {

	req := VolCreateRequest{}
	req.StripeCount = 2

	v := NewVolumeEntry(req)
	tests.Assert(t, v.StripeCount == 2)
}

func TestNewVolumeEntryFromRequestDisperse(t *testing.T) {

	req := VolCreateRequest{}
	req.DisperseCount = 2

	v := NewVolumeEntry(req)
	tests.Assert(t, v.DisperseCount == 2)
}

func TestNewVolumeEntryFromRequestRedundancy(t *testing.T) {

	req := VolCreateRequest{}
	req.RedundancyCount = 2

	v := NewVolumeEntry(req)
	tests.Assert(t, v.RedundancyCount == 2)
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

	req := VolCreateRequest{}
	req.Bricks = []string{"abc", "def"}

	v := NewVolumeEntry(req)
	tests.Assert(t, find(v.Bricks, "abc"))
	tests.Assert(t, find(v.Bricks, "def"))
}
