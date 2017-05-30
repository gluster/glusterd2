package volume

import (
	"fmt"
	"os"
	"testing"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/tests"

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

// TestNewBrickEntry validates NewBrickEntries ()
func TestNewBrickEntry(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()

	bricks := getSampleBricks("/tmp/b1", "/tmp/b2")
	brickPaths := []string{"/tmp/b1", "/tmp/b2"}
	host, _ := os.Hostname()

	b, err := NewBrickEntriesFunc(bricks, "volume")
	tests.Assert(t, err == nil)
	tests.Assert(t, b != nil)
	for _, brick := range b {
		tests.Assert(t, find(brickPaths, brick.Path))
		tests.Assert(t, host == brick.Hostname)
	}

	// Some negative tests
	mockBricks := []string{"/tmp/b1", "/tmp/b2"} //with out IPs
	_, err = NewBrickEntriesFunc(mockBricks, "volume")
	tests.Assert(t, err != nil)

	//Now mock filepath.Abs()
	defer heketitests.Patch(&absFilePath, func(path string) (string, error) {
		return "", errors.ErrBrickPathConvertFail
	}).Restore()

	_, err = NewBrickEntriesFunc(bricks, "volume")
	tests.Assert(t, err == errors.ErrBrickPathConvertFail)

}
