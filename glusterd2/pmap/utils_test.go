package pmap

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPortFree(t *testing.T) {

	assert := require.New(t)

	// listen on a random port
	l, err := net.Listen("tcp", ":0")
	assert.NoError(err)
	defer l.Close()

	_, portStr, err := net.SplitHostPort(l.Addr().String())
	assert.NoError(err)

	port, err := strconv.Atoi(portStr)
	assert.NoError(err)

	assert.False(isPortFree(port))
	l.Close()
	assert.True(isPortFree(port))
}
