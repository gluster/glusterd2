package utils

import (
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"testing"

	"github.com/gluster/glusterd2/tests"

	"github.com/pborman/uuid"
)

func TestIsLocalAddress(t *testing.T) {
	host, _ := os.Hostname()
	var local bool
	var e error
	local, e = IsLocalAddress(host)
	tests.Assert(t, e == nil)
	tests.Assert(t, local == true)

	local, e = IsLocalAddress("invalid ip")
	tests.Assert(t, local == false)
	tests.Assert(t, e != nil)

	local, e = IsLocalAddress("127.0.0.1")
	tests.Assert(t, local == true)
	tests.Assert(t, e == nil)

	local, e = IsLocalAddress("122.122.122.122.122")
	tests.Assert(t, local == false)
	tests.Assert(t, e != nil)
}

func TestParseHostAndBrickPath(t *testing.T) {
	hostname := "abc"
	brick := "/brick"
	brickPath := "abc:/brick"
	var h, b string
	var e error

	h, b, e = ParseHostAndBrickPath(brickPath)
	tests.Assert(t, e == nil)
	tests.Assert(t, h == hostname)
	tests.Assert(t, b == brick)

	h, b, e = ParseHostAndBrickPath("invalid brick")
	tests.Assert(t, e != nil)
	tests.Assert(t, len(h) == 0)
	tests.Assert(t, len(b) == 0)

	h, b, e = ParseHostAndBrickPath("a:b:c")
	tests.Assert(t, e == nil)
	tests.Assert(t, h == "a:b")
	tests.Assert(t, b == "c")
}

func TestValidateBrickPathLength(t *testing.T) {
	var brick string
	for i := 0; i <= unix.PathMax; i++ {
		brick = brick + "a"
	}
	tests.Assert(t, ValidateBrickPathLength(brick) != nil)
	tests.Assert(t, ValidateBrickPathLength("/brick/b1") == nil)
}

func TestValidateBrickSubDirLength(t *testing.T) {
	brick := "/tmp/"
	for i := 0; i <= PosixPathMax; i++ {
		brick = brick + "a"
	}
	tests.Assert(t, ValidateBrickSubDirLength(brick) != nil)
	tests.Assert(t, ValidateBrickSubDirLength("/tmp/brick1") == nil)
}

func TestValidateBrickPathStats(t *testing.T) {
	tests.Assert(t, ValidateBrickPathStats("/b1", "host", false) != nil)
	tests.Assert(t, ValidateBrickPathStats("/b1", "host", true) == nil)
	tests.Assert(t, ValidateBrickPathStats("/tmp", "host", false) != nil)
	//TODO : In build system /tmp is considered as root, hence passing
	//force = true
	tests.Assert(t, ValidateBrickPathStats("/tmp/b1", "host", true) == nil)
	cmd := exec.Command("touch", "/tmp/b1/b2")
	err := cmd.Run()
	tests.Assert(t, err == nil)
	tests.Assert(t, ValidateBrickPathStats("/tmp/b1/b2", "host", false) != nil)
}

func TestValidateXattrSupport(t *testing.T) {
	defer tests.Patch(&Setxattr, tests.MockSetxattr).Restore()
	defer tests.Patch(&Getxattr, tests.MockGetxattr).Restore()
	defer tests.Patch(&Removexattr, tests.MockRemovexattr).Restore()
	tests.Assert(t, ValidateXattrSupport("/tmp/b1", "localhost", uuid.NewRandom(), true) == nil)
}
