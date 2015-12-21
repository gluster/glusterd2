package volume

import (
	"fmt"
	"os"
	"testing"

	"github.com/gluster/glusterd2/tests"
	"github.com/gluster/glusterd2/utils"
)

func find(haystack []string, needle string) bool {

	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}

	return false
}

func mockGetVolumes() ([]Volinfo, error) {
	return nil, nil
}

// getSampleBricks prepare a list of couple of bricks with the path names as
// input along with the local hostname
func getSampleBricks(b1 string, b2 string) []string {

	var bricks []string
	lhost, _ := os.Hostname()
	brick1 := fmt.Sprintf("%s:%s", lhost, b1)
	brick2 := fmt.Sprintf("%s:%s", lhost, b2)
	bricks = append(bricks, brick1)
	bricks = append(bricks, brick2)
	return bricks
}

// TestNewVolumeEntry tests whether the volinfo object is successfully created
func TestNewVolumeObject(t *testing.T) {
	v := NewVolinfo()

	tests.Assert(t, v.Options != nil)
	tests.Assert(t, len(v.ID) == 0)
}

// TestNewBrickEntryFromRequestBricksRootPartition checks whether bricks can be
// created from root partition with a force option
func TestNewBrickEntryFromRequestBricksRootPartition(t *testing.T) {
	bricks := getSampleBricks("/b1", "/b2")

	b, err := NewBrickEntries(bricks)
	tests.Assert(t, err == nil)
	tests.Assert(t, b != nil)

}

// TestNewBrickEntryFromRequestBricks checks if bricks are successfully created
// from the request
func TestNewBrickEntryFromRequestBricks(t *testing.T) {
	bricks := getSampleBricks("/tmp/b1", "/tmp/b2")
	brickPaths := []string{"/tmp/b1", "/tmp/b2"}
	host, _ := os.Hostname()

	b, err := NewBrickEntries(bricks)
	tests.Assert(t, err == nil)
	tests.Assert(t, b != nil)
	for _, brick := range b {
		tests.Assert(t, find(brickPaths, brick.Path))
		tests.Assert(t, host == brick.Hostname)
	}

}

// TestNewVolumeEntryFromRequest tests whether the volume is created with a
// valid request
func TestNewVolumeEntryFromRequest(t *testing.T) {
	var err error
	defer tests.Patch(&utils.Setxattr, tests.MockSetxattr).Restore()
	defer tests.Patch(&utils.Getxattr, tests.MockGetxattr).Restore()
	defer tests.Patch(&utils.Removexattr, tests.MockRemovexattr).Restore()
	defer tests.Patch(&MGetVolumes, mockGetVolumes).Restore()

	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	req.Force = true
	v, e := NewVolumeEntry(req)
	tests.Assert(t, e == nil)
	tests.Assert(t, v.Name == "vol1")
	tests.Assert(t, v.Transport == "tcp")
	tests.Assert(t, v.ReplicaCount == 1)
	tests.Assert(t, len(v.ID) != 0)
	v.Bricks, err = NewBrickEntries(req.Bricks)
	tests.Assert(t, err == nil)
	tests.Assert(t, v.Bricks != nil)
	tests.Assert(t, len(v.Bricks) != 0)
	_, err = ValidateBrickEntries(v.Bricks, v.ID, true)
	tests.Assert(t, err == nil)
	defer tests.Patch(&MValidateBrickPathStats, tests.MockValidateBrickPathStats).Restore()
	_, err = ValidateBrickEntries(v.Bricks, v.ID, false)
	tests.Assert(t, err == nil)

}

// TestNewVolumeEntryFromRequestReplica validates whether the volume create is
// successful with given replica information
func TestNewVolumeEntryFromRequestReplica(t *testing.T) {
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	req.Force = true
	req.ReplicaCount = 3

	v, _ := NewVolumeEntry(req)
	tests.Assert(t, v.ReplicaCount == 3)
}

// TestNewVolumeEntryFromRequestTransport validates whether the volume create is
// successful with given transport type
func TestNewVolumeEntryFromRequestTransport(t *testing.T) {
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Transport = "rdma"
	req.Force = true
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	v, _ := NewVolumeEntry(req)
	tests.Assert(t, v.Transport == "rdma")
}

// TestNewVolumeEntryFromRequestStripe validates whether the volume create is
// successful with given stripe count
func TestNewVolumeEntryFromRequestStripe(t *testing.T) {
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	req.Force = true
	req.StripeCount = 2

	v, _ := NewVolumeEntry(req)
	tests.Assert(t, v.StripeCount == 2)
}

// TestNewVolumeEntryFromRequestDisperse validates whether the volume create is
// successful with given disperse count
func TestNewVolumeEntryFromRequestDisperse(t *testing.T) {
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Force = true
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	req.DisperseCount = 2

	v, _ := NewVolumeEntry(req)
	tests.Assert(t, v.DisperseCount == 2)
}

// TestNewVolumeEntryFromRequestRedundancy validates whether the volume create
// is successful with given redundancy count
func TestNewVolumeEntryFromRequestRedundancy(t *testing.T) {
	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Force = true
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	req.RedundancyCount = 2
	//TODO : This test needs improvement as redundancy count is tightly
	//coupled with disperse count, ideally this should fail
	v, _ := NewVolumeEntry(req)
	tests.Assert(t, v.RedundancyCount == 2)
}
