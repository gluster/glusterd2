package volumecommands

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/gluster/glusterd2/brick"
	gderrors "github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/tests"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	"github.com/pborman/uuid"

	heketitests "github.com/heketi/tests"
)

var (
	errBad = errors.New("bad")
)

//TestUnmarshalVolCreateRequest validates the JSON request of volume
//create request
func TestUnmarshalVolCreateRequest(t *testing.T) {
	msg := new(VolCreateRequest)
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

	msg := new(VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	vol, e := createVolinfo(msg)
	tests.Assert(t, e == nil && vol != nil)

	// Mock failure in NewBrickEntries(), createVolume() should fail
	defer heketitests.Patch(&volume.NewBrickEntriesFunc, func(bricks []string, volName string) ([]brick.Brickinfo, error) {
		return nil, errBad
	}).Restore()
	_, e = createVolinfo(msg)
	tests.Assert(t, e == errBad)
}

// TestValidateVolumeCreate validates validateVolumeCreate()
func TestValidateVolumeCreate(t *testing.T) {
	msg := new(VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}

	c := transaction.NewMockCtx()
	c.Set("req", msg)

	defer heketitests.Patch(&volume.ValidateBrickEntriesFunc, func(bricks []brick.Brickinfo, volID uuid.UUID, force bool) (int, error) {
		return 0, nil
	}).Restore()
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()

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

// TestGenerateVolfiles validates generateVolfiles
func TestGenerateVolfiles(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	msg := new(VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}

	vol, e := createVolinfo(msg)

	c := transaction.NewMockCtx()
	c.Set("volinfo", vol)

	fakeVolauth := volume.VolAuth{
		Username: uuid.NewRandom().String(),
		Password: uuid.NewRandom().String(),
	}
	c.Set("volauth", fakeVolauth)

	defer heketitests.Patch(&volgen.GenerateVolfileFunc, func(vinfo *volume.Volinfo, vauth *volume.VolAuth) error {
		return nil
	}).Restore()
	defer heketitests.Patch(&volume.AddOrUpdateVolumeFunc, func(vinfo *volume.Volinfo) error {
		return nil
	}).Restore()

	e = generateVolfiles(c)
	tests.Assert(t, e == nil)

	// Mock volgen failure
	defer heketitests.Patch(&volgen.GenerateVolfileFunc, func(vinfo *volume.Volinfo, vauth *volume.VolAuth) error {
		return errBad
	}).Restore()
	e = generateVolfiles(c)
	tests.Assert(t, e == errBad)

	defer heketitests.Patch(&volgen.GenerateVolfileFunc, func(vinfo *volume.Volinfo, vauth *volume.VolAuth) error {
		return nil
	}).Restore()
}

// TestStoreVolume tests storeVolume
func TestStoreVolume(t *testing.T) {
	defer heketitests.Patch(&peer.GetPeerIDByAddrF, peer.GetPeerIDByAddrMockGood).Restore()
	msg := new(VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}

	vol, e := createVolinfo(msg)

	c := transaction.NewMockCtx()
	c.Set("volinfo", vol)
	// Mock store failure
	defer heketitests.Patch(&volume.AddOrUpdateVolumeFunc, func(vinfo *volume.Volinfo) error {
		return errBad
	}).Restore()
	e = storeVolume(c)
	tests.Assert(t, e == errBad)

}
