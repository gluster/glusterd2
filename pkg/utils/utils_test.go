package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
