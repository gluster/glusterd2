package volumecommands

import (
	"errors"
	"testing"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/testutils"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	errBad = errors.New("bad")
)

// TestCreateVolinfo validates newVolinfo()
func TestCreateVolinfo(t *testing.T) {
	defer testutils.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	defer testutils.Patch(&peer.GetPeerF, peer.GetPeerFMockGood).Restore()

	msg := new(api.VolCreateReq)
	u := uuid.NewRandom()
	msg.Name = "vol"
	msg.Subvols = []api.SubvolReq{{Bricks: []api.BrickReq{
		{PeerID: u.String(), Path: "/tmp/b1"},
		{PeerID: u.String(), Path: "/tmp/b2"},
	}}}
	msg.Metadata = make(map[string]string)
	msg.Metadata["owner"] = "gd2test"
	vol, e := newVolinfo(msg)
	assert.Nil(t, e)
	assert.NotNil(t, vol)

	// Mock failure in NewBrickEntries(), createVolume() should fail
	defer testutils.Patch(&volume.NewBrickEntriesFunc, func(bricks []api.BrickReq, volName string, volID uuid.UUID) ([]brick.Brickinfo, error) {
		return nil, errBad
	}).Restore()
	_, e = newVolinfo(msg)
	assert.Equal(t, errBad, e)
}
