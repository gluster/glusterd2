package pmap

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	regType := &registryType{}
	registryBind(1000, "/tmp/brick", 0, nil)

	res, err := regType.MarshalJSON()
	assert.NotNil(t, res)
	assert.Nil(t, err)

	resString := regType.String()
	assert.NotEmpty(t, resString)
}

func TestIsPortFree(t *testing.T) {
	port := 49686
	server, err := openport(port)

	if err == nil {
		res := isPortFree(port)
		assert.False(t, res)

		server.Close()
		res = isPortFree(port)
		assert.True(t, res)
	}

}

func TestStringInSlice(t *testing.T) {
	res := stringInSlice("test1", []string{"test1", "test2"})
	assert.True(t, res)

	res = stringInSlice("test", []string{"test1", "test2"})
	assert.False(t, res)

}

func openport(port int) (net.Listener, error) {

	host := ":" + strconv.Itoa(port)
	server, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}
	return server, nil
}
