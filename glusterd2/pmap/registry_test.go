package pmap

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {

	assert := require.New(t)

	r := &pmapRegistry{
		basePort:    1000,
		bricks:      make(map[string]int),
		conns:       make(map[net.Conn]int),
		portLockFds: make(map[int]int),
	}

	for i := 1000; i <= (r.basePort + 100); i++ {
		r.ports[i].State = PortLeased
	}

	// test sign in path
	for i := 1000; i <= (r.basePort + 100); i++ {
		err := r.Update(i, fmt.Sprintf("/tmp/brick%d", i), nil)
		assert.NoError(err)
		assert.Equal(r.ports[i].State, PortInUse)
	}

	for _, v := range r.bricks {
		assert.NotZero(v)
	}

	err := r.Update(math.MaxInt32, "some_brick", nil)
	assert.Error(err)

	// test port search
	for i := 1000; i <= (r.basePort + 100); i++ {
		p, err := r.SearchByBrickPath(fmt.Sprintf("/tmp/brick%d", i))
		assert.NoError(err)
		assert.Equal(p, i)
	}

	p, err := r.SearchByBrickPath("non-existent-brick")
	assert.Error(err)
	assert.Equal(p, -1)

	p, err = r.SearchByBrickPath("")
	assert.Error(err)
	assert.Equal(p, -1)

	// test sign out path
	for i := 1000; i <= (r.basePort + 100); i++ {
		bpath := fmt.Sprintf("/tmp/brick%d", i)
		err := r.Remove(i, bpath, nil)
		assert.NoError(err)
		p, err := r.SearchByBrickPath(bpath)
		assert.Error(err)
		assert.Equal(p, -1)
	}

	err = r.Remove(math.MaxInt32, "some_brick", nil)
	assert.Error(err)
}
