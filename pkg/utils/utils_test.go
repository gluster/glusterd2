package utils

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/gluster/glusterd2/pkg/testutils"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestIsLocalAddress(t *testing.T) {
	host, _ := os.Hostname()
	var local bool
	var e error
	local, e = IsLocalAddress(host)
	assert.Nil(t, e)
	assert.True(t, local)

	local, e = IsLocalAddress("invalid ip")
	assert.NotNil(t, e)
	assert.False(t, local)

	local, e = IsLocalAddress("127.0.0.1")
	assert.Nil(t, e)
	assert.True(t, local)

	local, e = IsLocalAddress("122.122.122.122.122")
	assert.NotNil(t, e)
	assert.False(t, local)
}

func TestParseHostAndBrickPath(t *testing.T) {
	brick := "/brick"
	brickPath := "abc:/brick"
	var h, b string
	var e error

	h, b, e = ParseHostAndBrickPath(brickPath)
	assert.Nil(t, e)
	assert.Equal(t, brick, b)

	h, b, e = ParseHostAndBrickPath("invalid brick")
	assert.NotNil(t, e)
	assert.Empty(t, h)
	assert.Empty(t, b)

	h, b, e = ParseHostAndBrickPath("a:b:c")
	assert.Nil(t, e)
	assert.Equal(t, "a:b", h)
	assert.Equal(t, "c", b)
}

func TestValidateBrickPathLength(t *testing.T) {
	var brick string
	for i := 0; i <= unix.PathMax; i++ {
		brick = brick + "a"
	}
	assert.NotNil(t, ValidateBrickPathLength(brick))
	assert.Nil(t, ValidateBrickPathLength("/brick/b1"))
}

func TestValidateBrickSubDirLength(t *testing.T) {
	brick := "/tmp/"
	for i := 0; i <= PosixPathMax; i++ {
		brick = brick + "a"
	}
	assert.NotNil(t, ValidateBrickSubDirLength(brick))
	assert.Nil(t, ValidateBrickSubDirLength("/tmp/brick1"))
}

func TestValidateBrickPathStats(t *testing.T) {
	assert.NotNil(t, ValidateBrickPathStats("/bricks/b1", false))
	assert.Nil(t, ValidateBrickPathStats("/bricks/b1", true))
	assert.NotNil(t, ValidateBrickPathStats("/tmp", false))
	//TODO : In build system /tmp is considered as root, hence passing
	//force = true
	assert.Nil(t, ValidateBrickPathStats("/tmp/bricks/b1", true))
	cmd := exec.Command("touch", "/tmp/bricks/b1/b2")
	err := cmd.Run()
	assert.Nil(t, err)
	assert.NotNil(t, ValidateBrickPathStats("/tmp/bricks/b1/b2", false))
}

func TestValidateXattrSupport(t *testing.T) {
	defer testutils.Patch(&Setxattr, testutils.MockSetxattr).Restore()
	defer testutils.Patch(&Getxattr, testutils.MockGetxattr).Restore()
	defer testutils.Patch(&Removexattr, testutils.MockRemovexattr).Restore()
	assert.Nil(t, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true))

	// Some negative tests
	var xattrErr error
	baderror := errors.New("Bad")
	xattrErr = baderror

	// Now check what happens when setxattr fails
	defer testutils.Patch(&Setxattr, func(path string, attr string, data []byte, flags int) (err error) {
		return xattrErr
	}).Restore()
	assert.Equal(t, baderror, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true))

	// Now check what happens when getxattr fails
	defer testutils.Patch(&Getxattr, func(path string, attr string, dest []byte) (sz int, err error) {
		return 0, xattrErr
	}).Restore()
	assert.Equal(t, baderror, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true))

	// Now check what happens when removexattr fails
	defer testutils.Patch(&Removexattr, func(path string, attr string) (err error) {
		return xattrErr
	}).Restore()
	assert.Equal(t, baderror, ValidateXattrSupport("/tmp/b1", uuid.NewRandom(), true))

}
