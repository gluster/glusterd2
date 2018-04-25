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

func TestInitDir(t *testing.T) {
	validPath := "/tmp/initdir"
	filePath := "/tmp/abc.txt"
	defer os.Remove(validPath)

	err := InitDir(validPath)
	assert.Nil(t, err)

	os.Create(filePath)
	defer os.Remove(filePath)

	err = InitDir(filePath)
	assert.Contains(t, err.Error(), "not a directory")
}
func testfuncname() {

}
func TestGetFuncName(t *testing.T) {
	name := GetFuncName(testfuncname)
	assert.Contains(t, name, "utils.testfuncname")
}

func TestStringInSlice(t *testing.T) {
	resp := StringInSlice("paas", []string{"paas", "testing"})
	assert.True(t, resp)

	resp = StringInSlice("fail", []string{"paas", "testing"})
	assert.False(t, resp)
}

func TestIsAddressSame(t *testing.T) {
	resp := IsAddressSame("192.168.1.1", "192.168.1.1")
	assert.True(t, resp)

	resp = IsAddressSame("192.169.1.2", "192.168.1.1")
	assert.False(t, resp)

	resp = IsAddressSame("192.169.1.2", "192.168.1")
	assert.False(t, resp)

	resp = IsAddressSame("192.169.1", "192.168.1.1")
	assert.False(t, resp)

}

func TestGetLocalIP(t *testing.T) {
	ip, err := GetLocalIP()
	assert.Nil(t, err)
	assert.NotEmpty(t, ip)
}
