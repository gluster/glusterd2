package volume

import (
	"fmt"
	"testing"

	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/testutils"

	heketitests "github.com/heketi/tests"
	"github.com/pborman/uuid"
)

func find(haystack []string, needle string) bool {

	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}

	return false
}

// getSampleBricks prepare a list of couple of bricks with the path names as
// input along with the local uuid
func getSampleBricks(b1 string, b2 string) []string {

	var bricks []string
	lhost := uuid.NewRandom()
	brick1 := fmt.Sprintf("%s:%s", lhost, b1)
	brick2 := fmt.Sprintf("%s:%s", lhost, b2)
	bricks = append(bricks, brick1)
	bricks = append(bricks, brick2)
	return bricks
}

// TestNewBrickEntry validates NewBrickEntries ()
func TestNewBrickEntry(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	defer heketitests.Patch(&peer.GetPeerF, peer.GetPeerFMockGood).Restore()

	bricks := getSampleBricks("/tmp/b1", "/tmp/b2")
	brickPaths := []string{"/tmp/b1", "/tmp/b2"}

	b, err := NewBrickEntriesFunc(bricks, "volume", nil)
	testutils.Assert(t, err == nil)
	testutils.Assert(t, b != nil)
	for _, brick := range b {
		testutils.Assert(t, find(brickPaths, brick.Path))
	}

	// Some negative tests
	mockBricks := []string{"/tmp/b1", "/tmp/b2"} //with out IPs
	_, err = NewBrickEntriesFunc(mockBricks, "volume", nil)
	testutils.Assert(t, err != nil)

	//Now mock filepath.Abs()
	defer heketitests.Patch(&absFilePath, func(path string) (string, error) {
		return "", errors.ErrBrickPathConvertFail
	}).Restore()

	_, err = NewBrickEntriesFunc(bricks, "volume", nil)
	testutils.Assert(t, err == errors.ErrBrickPathConvertFail)

}
