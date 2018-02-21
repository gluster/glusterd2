package volumecommands

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/testutils"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	errBad = errors.New("bad")
)

//TestUnmarshalVolCreateRequest validates the JSON request of volume
//create request
func TestUnmarshalVolCreateRequest(t *testing.T) {
	msg := new(api.VolCreateReq)
	assert.NotNil(t, msg)

	// Request with invalid JSON format
	r, _ := http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"invalid_format"}`)))
	_, e := unmarshalVolCreateRequest(msg, r)
	assert.Equal(t, gderrors.ErrJSONParsingFailed, e)

	// Request with empty volume name
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{}`)))
	_, e = unmarshalVolCreateRequest(msg, r)
	assert.Equal(t, gderrors.ErrEmptyVolName, e)

	// Request with empty bricks
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"name" : "vol"}`)))
	_, e = unmarshalVolCreateRequest(msg, r)
	assert.Equal(t, "vol", msg.Name)
	assert.Equal(t, gderrors.ErrEmptyBrickList, e)

	// Request with volume name & bricks
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"name" : "vol", "subvols": [{"bricks":[{"nodeid": "127.0.0.1", "path": "/tmp/b1"}]}]}`)))
	_, e = unmarshalVolCreateRequest(msg, r)
	assert.Nil(t, e)

}

// TestCreateVolinfo validates createVolinfo()
func TestCreateVolinfo(t *testing.T) {
	defer testutils.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	defer testutils.Patch(&peer.GetPeerF, peer.GetPeerFMockGood).Restore()

	msg := new(api.VolCreateReq)
	u := uuid.NewRandom()
	msg.Name = "vol"
	msg.Subvols = []api.SubvolReq{{Bricks: []api.BrickReq{
		{NodeID: u.String(), Path: "/tmp/b1"},
		{NodeID: u.String(), Path: "/tmp/b2"},
	}}}
	vol, e := createVolinfo(msg)
	assert.Nil(t, e)
	assert.NotNil(t, vol)

	// Mock failure in NewBrickEntries(), createVolume() should fail
	defer testutils.Patch(&volume.NewBrickEntriesFunc, func(bricks []api.BrickReq, volName string, volID uuid.UUID) ([]brick.Brickinfo, error) {
		return nil, errBad
	}).Restore()
	_, e = createVolinfo(msg)
	assert.Equal(t, errBad, e)
}
