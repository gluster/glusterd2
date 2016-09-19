package volume

import (
	"fmt"
	"os"
	"testing"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/tests"
	"github.com/gluster/glusterd2/utils"

	heketitests "github.com/heketi/tests"
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
	v := NewVolinfoFunc()

	tests.Assert(t, v.Options != nil)
	tests.Assert(t, len(v.ID) == 0)

	// Negative test
	defer heketitests.Patch(&NewVolinfoFunc, func() (vol *Volinfo) {
		return nil
	}).Restore()
	v1 := NewVolinfoFunc()
	tests.Assert(t, v1 == nil)
}

// TestNewVolumeEntry validates NewVolumeEntry()
func TestNewVolumeEntry(t *testing.T) {
	req := new(VolCreateRequest)
	v, e := NewVolumeEntry(req)
	tests.Assert(t, e == nil)
	tests.Assert(t, v != nil)

	// Negative test - mock out NewVolInfo()
	defer heketitests.Patch(&NewVolinfoFunc, func() (vol *Volinfo) {
		return nil
	}).Restore()

	_, e = NewVolumeEntry(req)
	tests.Assert(t, e == errors.ErrVolCreateFail)
}

// TestNewBrickEntry validates NewBrickEntries ()
func TestNewBrickEntry(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()

	bricks := getSampleBricks("/tmp/b1", "/tmp/b2")
	brickPaths := []string{"/tmp/b1", "/tmp/b2"}
	host, _ := os.Hostname()

	b, err := NewBrickEntriesFunc(bricks)
	tests.Assert(t, err == nil)
	tests.Assert(t, b != nil)
	for _, brick := range b {
		tests.Assert(t, find(brickPaths, brick.Path))
		tests.Assert(t, host == brick.Hostname)
	}

	// Some negative tests
	mockBricks := []string{"/tmp/b1", "/tmp/b2"} //with out IPs
	_, err = NewBrickEntriesFunc(mockBricks)
	tests.Assert(t, err != nil)

	//Now mock filepath.Abs()
	defer heketitests.Patch(&absFilePath, func(path string) (string, error) {
		return "", errors.ErrBrickPathConvertFail
	}).Restore()

	_, err = NewBrickEntriesFunc(bricks)
	tests.Assert(t, err == errors.ErrBrickPathConvertFail)

}

// TestNewVolumeEntryFromRequest tests whether the volume is created with a
// valid request
func TestNewVolumeEntryFromRequest(t *testing.T) {
	var err error
	defer heketitests.Patch(&utils.PathMax, 4096).Restore()
	defer heketitests.Patch(&utils.Setxattr, tests.MockSetxattr).Restore()
	defer heketitests.Patch(&utils.Getxattr, tests.MockGetxattr).Restore()
	defer heketitests.Patch(&utils.Removexattr, tests.MockRemovexattr).Restore()
	defer heketitests.Patch(&getVolumesFunc, mockGetVolumes).Restore()

	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	//peer.GetPeerIDByAddrF = peer.GetPeerIDByAddrMockGood
	//defer func() { peer.GetPeerIDByAddrF = peer.GetPeerIDByAddr }()

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
	v.Bricks, err = NewBrickEntriesFunc(req.Bricks)
	tests.Assert(t, err == nil)
	tests.Assert(t, v.Bricks != nil)
	tests.Assert(t, len(v.Bricks) != 0)
	_, err = ValidateBrickEntriesFunc(v.Bricks, v.ID, true)
	tests.Assert(t, err == nil)
	defer heketitests.Patch(&validateBrickPathStatsFunc, tests.MockValidateBrickPathStats).Restore()
	_, err = ValidateBrickEntriesFunc(v.Bricks, v.ID, false)
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

func TestRemoveBrickPaths(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()

	req := new(VolCreateRequest)
	req.Name = "vol1"
	req.Bricks = getSampleBricks("/tmp/b1", "/tmp/b2")
	v, e := NewVolumeEntry(req)
	v.Bricks, e = NewBrickEntriesFunc(req.Bricks)
	e = RemoveBrickPaths(v.Bricks)
	tests.Assert(t, e == nil)
}
