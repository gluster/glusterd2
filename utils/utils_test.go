package utils

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/gluster/glusterd2/tests"
)

func TestIsLocalAddress(t *testing.T) {
	host, _ := os.Hostname()
	var local bool
	var e error
	local, e = IsLocalAddress(host)
	tests.Assert(t, local == true)
	tests.Assert(t, e == nil)

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

	h, b = ParseHostAndBrickPath(brickPath)
	tests.Assert(t, h == hostname)
	tests.Assert(t, b == brick)

	h, b = ParseHostAndBrickPath("invalid brick")
	tests.Assert(t, len(h) == 0)
	tests.Assert(t, len(b) == 0)

	h, b = ParseHostAndBrickPath("a:b:c")
	tests.Assert(t, h == "a:b")
	tests.Assert(t, b == "c")
}

func TestValidateBrickPathLength(t *testing.T) {
	var brick string
	for i := 0; i <= syscall.PathMax; i++ {
		brick = brick + "a"
	}
	tests.Assert(t, ValidateBrickPathLength(brick) != 0)
	tests.Assert(t, ValidateBrickPathLength("/brick/b1") == 0)
}

func TestValidateBrickSubDirLength(t *testing.T) {
	brick := "/tmp/"
	for i := 0; i <= PosixPathMax; i++ {
		brick = brick + "a"
	}
	tests.Assert(t, ValidateBrickSubDirLength(brick) != 0)
	tests.Assert(t, ValidateBrickSubDirLength("/tmp/brick1") == 0)
}

func TestValidateBrickPathStats(t *testing.T) {
	tests.Assert(t, ValidateBrickPathStats("/b1", "host", false) != nil)
	tests.Assert(t, ValidateBrickPathStats("/b1", "host", true) == nil)
	tests.Assert(t, ValidateBrickPathStats("/tmp", "host", false) != nil)
	tests.Assert(t, ValidateBrickPathStats("/tmp/b1", "host", false) == nil)
	cmd := exec.Command("touch", "/tmp/b1/b2")
	err := cmd.Run()
	tests.Assert(t, err == nil)
	tests.Assert(t, ValidateBrickPathStats("/tmp/b1/b2", "host", false) != nil)
}
