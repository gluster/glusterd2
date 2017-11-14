package volumecommands

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/gluster/glusterd2/bin/glusterd2/brick"
	"github.com/gluster/glusterd2/bin/glusterd2/peer"
	"github.com/gluster/glusterd2/bin/glusterd2/transaction"
	"github.com/gluster/glusterd2/bin/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/tests"

	"github.com/pborman/uuid"

	heketitests "github.com/heketi/tests"
)

var (
	errBad = errors.New("bad")
)

//TestUnmarshalVolCreateRequest validates the JSON request of volume
//create request
func TestUnmarshalVolCreateRequest(t *testing.T) {
	msg := new(api.VolCreateReq)
	tests.Assert(t, msg != nil)

	// Request with invalid JSON format
	r, _ := http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"invalid_format"}`)))
	_, e := unmarshalVolCreateRequest(msg, r)
	tests.Assert(t, e == gderrors.ErrJSONParsingFailed)

	// Request with empty volume name
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{}`)))
	_, e = unmarshalVolCreateRequest(msg, r)
	tests.Assert(t, e == gderrors.ErrEmptyVolName)

	// Request with empty bricks
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"name" : "vol"}`)))
	_, e = unmarshalVolCreateRequest(msg, r)
	tests.Assert(t, msg.Name == "vol")
	tests.Assert(t, e == gderrors.ErrEmptyBrickList)

	// Request with volume name & bricks
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"name" : "vol", "bricks":["127.0.0.1:/tmp/b1"]}`)))
	_, e = unmarshalVolCreateRequest(msg, r)
	tests.Assert(t, e == nil)

}

// TestCreateVolinfo validates createVolinfo()
func TestCreateVolinfo(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	defer heketitests.Patch(&peer.GetPeerF, peer.GetPeerFMockGood).Restore()

	msg := new(api.VolCreateReq)
	u := uuid.NewRandom()
	msg.Name = "vol"
	msg.Bricks = []string{u.String() + ":/tmp/b1", u.String() + ":/tmp/b2"}
	vol, e := createVolinfo(msg)
	tests.Assert(t, e == nil && vol != nil)

	// Mock failure in NewBrickEntries(), createVolume() should fail
	defer heketitests.Patch(&volume.NewBrickEntriesFunc, func(bricks []string, volName string, volID uuid.UUID) ([]brick.Brickinfo, error) {
		return nil, errBad
	}).Restore()
	_, e = createVolinfo(msg)
	tests.Assert(t, e == errBad)
}

// TestValidateVolumeCreate validates validateVolumeCreate()
func TestValidateVolumeCreate(t *testing.T) {
	msg := new(api.VolCreateReq)

	msg.Name = "vol"
	u := uuid.NewRandom()
	msg.Bricks = []string{u.String() + ":/tmp/b1", u.String() + ":/tmp/b2"}

	c := transaction.NewMockCtx()
	c.Set("req", msg)

	defer heketitests.Patch(&volume.ValidateBrickEntriesFunc, func(bricks []brick.Brickinfo, volID uuid.UUID, force bool) (int, error) {
		return 0, nil
	}).Restore()
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	defer heketitests.Patch(&peer.GetPeerF, peer.GetPeerFMockGood).Restore()

	vol, e := createVolinfo(msg)
	tests.Assert(t, e == nil)
	c.Set("volinfo", vol)

	e = validateVolumeCreate(c)
	tests.Assert(t, e == nil)

	// Mock validateBrickEntries failure
	defer heketitests.Patch(&volume.ValidateBrickEntriesFunc, func(bricks []brick.Brickinfo, volID uuid.UUID, force bool) (int, error) {
		return 0, errBad
	}).Restore()
	e = validateVolumeCreate(c)
	tests.Assert(t, e == errBad)
}
