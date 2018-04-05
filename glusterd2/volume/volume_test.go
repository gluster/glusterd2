package volume

import (
	"testing"

	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/testutils"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
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
func getSampleBricks(b1 string, b2 string) []api.BrickReq {

	lhost := uuid.NewRandom()
	return []api.BrickReq{
		{PeerID: lhost.String(), Path: b1},
		{PeerID: lhost.String(), Path: b2},
	}
}

// TestNewBrickEntry validates NewBrickEntries ()
func TestNewBrickEntry(t *testing.T) {
	defer testutils.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	defer testutils.Patch(&peer.GetPeerF, peer.GetPeerFMockGood).Restore()

	bricks := getSampleBricks("/tmp/b1", "/tmp/b2")
	brickPaths := []string{"/tmp/b1", "/tmp/b2"}

	b, err := NewBrickEntriesFunc(bricks, "volume", nil)
	assert.Nil(t, err)
	assert.NotNil(t, b)
	for _, brick := range b {
		assert.True(t, find(brickPaths, brick.Path))
	}

	// Some negative tests
	mockBricks := []api.BrickReq{
		{PeerID: "", Path: "/tmp/b1"},
		{PeerID: "", Path: "/tmp/b2"},
	} //with out IPs
	_, err = NewBrickEntriesFunc(mockBricks, "volume", nil)
	assert.NotNil(t, err)

	//Now mock filepath.Abs()
	defer testutils.Patch(&absFilePath, func(path string) (string, error) {
		return "", errors.ErrBrickPathConvertFail
	}).Restore()

	_, err = NewBrickEntriesFunc(bricks, "volume", nil)
	assert.Equal(t, errors.ErrBrickPathConvertFail, err)

}
