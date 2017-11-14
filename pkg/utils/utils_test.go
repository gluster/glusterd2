package utils

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/gluster/glusterd2/pkg/testutils"

	"github.com/pborman/uuid"

	heketitests "github.com/heketi/tests"
)

func TestIsLocalAddress(t *testing.T) {
	host, _ := os.Hostname()
	var local bool
	var e error
	local, e = IsLocalAddress(host)
	testutils.Assert(t, e == nil)
	testutils.Assert(t, local == true)

	local, e = IsLocalAddress("invalid ip")
	testutils.Assert(t, local == false)
	testutils.Assert(t, e != nil)

	local, e = IsLocalAddress("127.0.0.1")
	testutils.Assert(t, local == true)
	testutils.Assert(t, e == nil)

	local, e = IsLocalAddress("122.122.122.122.122")
	testutils.Assert(t, local == false)
	testutils.Assert(t, e != nil)
}

func TestParseHostAndBrickPath(t *testing.T) {
	brick := "/brick"
	brickPath := "abc:/brick"
	var h, b string
	var e error

	h, b, e = ParseHostAndBrickPath(brickPath)
	testutils.Assert(t, e == nil)
	testutils.Assert(t, b == brick)

	h, b, e = ParseHostAndBrickPath("invalid brick")
	testutils.Assert(t, e != nil)
	testutils.Assert(t, len(h) == 0)
	testutils.Assert(t, len(b) == 0)

	h, b, e = ParseHostAndBrickPath("a:b:c")
	testutils.Assert(t, e == nil)
	testutils.Assert(t, h == "a:b")
	testutils.Assert(t, b == "c")
}

func TestValidateBrickPathLength(t *testing.T) {
	var brick string
	for i := 0; i <= unix.PathMax; i++ {
		brick = brick + "a"
	}
	testutils.Assert(t, ValidateBrickPathLength(brick) != nil)
	testutils.Assert(t, ValidateBrickPathLength("/brick/b1") == nil)
}

func TestValidateBrickSubDirLength(t *testing.T) {
	brick := "/tmp/"
	for i := 0; i <= PosixPathMax; i++ {
		brick = brick + "a"
	}
	testutils.Assert(t, ValidateBrickSubDirLength(brick) != nil)
	testutils.Assert(t, ValidateBrickSubDirLength("/tmp/brick1") == nil)
}

func TestValidateBrickPathStats(t *testing.T) {
	testutils.Assert(t, ValidateBrickPathStats("/bricks/b1", false) != nil)
	testutils.Assert(t, ValidateBrickPathStats("/bricks/b1", true) == nil)
	testutils.Assert(t, ValidateBrickPathStats("/tmp", false) != nil)
	//TODO : In build system /tmp is considered as root, hence passing
	//force = true
	testutils.Assert(t, ValidateBrickPathStats("/tmp/bricks/b1", true) == nil)
	cmd := exec.Command("touch", "/tmp/bricks/b1/b2")
	err := cmd.Run()
	testutils.Assert(t, err == nil)
	testutils.Assert(t, ValidateBrickPathStats("/tmp/bricks/b1/b2", false) != nil)
}

func TestValidateXattrSupport(t *testing.T) {
	defer heketitests.Patch(&Setxattr, testutils.MockSetxattr).Restore()
	defer heketitests.Patch(&Getxattr, testutils.MockGetxattr).Restore()
	defer heketitests.Patch(&Removexattr, testutils.MockRemovexattr).Restore()
	testutils.Assert(t, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true) == nil)

	// Some negative tests
	var xattrErr error
	baderror := errors.New("Bad")
	xattrErr = baderror

	// Now check what happens when setxattr fails
	defer heketitests.Patch(&Setxattr, func(path string, attr string, data []byte, flags int) (err error) {
		return xattrErr
	}).Restore()
	testutils.Assert(t, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true) == baderror)

	// Now check what happens when getxattr fails
	defer heketitests.Patch(&Getxattr, func(path string, attr string, dest []byte) (sz int, err error) {
		return 0, xattrErr
	}).Restore()
	testutils.Assert(t, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true) == baderror)

	// Now check what happens when removexattr fails
	defer heketitests.Patch(&Removexattr, func(path string, attr string) (err error) {
		return xattrErr
	}).Restore()
	testutils.Assert(t, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true) == baderror)

}
