package volumecommands

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	gderrors "github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/tests"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	"github.com/pborman/uuid"

	heketitests "github.com/heketi/tests"
)

var (
	errBad = errors.New("bad")
)

//TestValidateVolumeCreateJSONRequest validates the JSON request of volume
//create request
func TestValidateVolumeCreateJSONRequest(t *testing.T) {
	msg := new(volume.VolCreateRequest)
	tests.Assert(t, msg != nil)

	// Request with invalid JSON format
	r, _ := http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"invalid_format"}`)))
	_, e := validateVolumeCreateJSONRequest(msg, r)
	tests.Assert(t, e == gderrors.ErrJSONParsingFailed)

	// Request with empty volume name
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{}`)))
	_, e = validateVolumeCreateJSONRequest(msg, r)
	tests.Assert(t, e == gderrors.ErrEmptyVolName)

	// Request with empty bricks
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"name" : "vol"}`)))
	_, e = validateVolumeCreateJSONRequest(msg, r)
	tests.Assert(t, msg.Name == "vol")
	tests.Assert(t, e == gderrors.ErrEmptyBrickList)

	// Request with volume name & bricks
	r, _ = http.NewRequest("POST", "/v1/volumes/", bytes.NewBuffer([]byte(`{"name" : "vol", "bricks":["127.0.0.1:/tmp/b1"]}`)))
	_, e = validateVolumeCreateJSONRequest(msg, r)
	tests.Assert(t, e == nil)

}

// TestCreateVolume validates createVolume()
func TestCreateVolume(t *testing.T) {
	msg := new(volume.VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	vol, e := createVolume(msg)
	tests.Assert(t, e == nil && vol != nil)

	// Mock failure in NewBrickEntries(), createVolume() should fail
	defer heketitests.Patch(&volume.NewBrickEntriesFunc, func(bricks []string) ([]volume.Brickinfo, error) {
		return nil, errBad
	}).Restore()
	vol, e = createVolume(msg)
	tests.Assert(t, e == errBad)
}

// TestValidateVolumeCreate validates validateVolumeCreate()
func TestValidateVolumeCreate(t *testing.T) {
	msg := new(volume.VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}
	vol, e := createVolume(msg)

	defer heketitests.Patch(&volume.ExistsFunc, func(name string) bool {
		return false
	}).Restore()
	defer heketitests.Patch(&volume.ValidateBrickEntriesFunc, func(bricks []volume.Brickinfo, volID uuid.UUID, force bool) (int, error) {
		return 0, nil
	}).Restore()

	_, e = validateVolumeCreate(msg, vol)
	tests.Assert(t, e == nil)

	// Mock volume exists failure
	defer heketitests.Patch(&volume.ExistsFunc, func(name string) bool {
		return true
	}).Restore()
	_, e = validateVolumeCreate(msg, vol)
	tests.Assert(t, e == gderrors.ErrVolExists)

	// Mock validateBrickEntries failure
	defer heketitests.Patch(&volume.ExistsFunc, func(name string) bool {
		return false
	}).Restore()

	defer heketitests.Patch(&volume.ValidateBrickEntriesFunc, func(bricks []volume.Brickinfo, volID uuid.UUID, force bool) (int, error) {
		return 0, errBad
	}).Restore()
	_, e = validateVolumeCreate(msg, vol)
	tests.Assert(t, e == errBad)
}

// TestCommitVolumeCreate validates commitVolumeCreate()
func TestCommitVolumeCreate(t *testing.T) {
	msg := new(volume.VolCreateRequest)

	msg.Name = "vol"
	msg.Bricks = []string{"127.0.0.1:/tmp/b1", "127.0.0.1:/tmp/b2"}

	vol, e := createVolume(msg)

	defer heketitests.Patch(&volgen.GenerateVolfileFunc, func(vinfo *volume.Volinfo) error {
		return nil
	}).Restore()
	defer heketitests.Patch(&volume.AddOrUpdateVolumeFunc, func(vinfo *volume.Volinfo) error {
		return nil
	}).Restore()

	_, e = commitVolumeCreate(vol)
	tests.Assert(t, e == nil)

	// Mock volgen failure
	defer heketitests.Patch(&volgen.GenerateVolfileFunc, func(vinfo *volume.Volinfo) error {
		return errBad
	}).Restore()
	_, e = commitVolumeCreate(vol)
	tests.Assert(t, e == errBad)

	defer heketitests.Patch(&volgen.GenerateVolfileFunc, func(vinfo *volume.Volinfo) error {
		return nil
	}).Restore()

	// Mock store failure
	defer heketitests.Patch(&volume.AddOrUpdateVolumeFunc, func(vinfo *volume.Volinfo) error {
		return errBad
	}).Restore()
	_, e = commitVolumeCreate(vol)
	tests.Assert(t, e == errBad)

}
