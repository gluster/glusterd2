package pmap

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {

	assert := require.New(t)

	r := &pmapRegistry{
		basePort:    1000,
		bricks:      make(map[string]int),
		conns:       make(map[net.Conn]int),
		portLockFds: make(map[int]int),
	}

	for i := 1000; i <= (r.basePort + 10); i++ {
		if i%2 == 0 {
			r.ports[i].State = PortInUse
			r.Update(i, fmt.Sprintf("/tmp/brick%d", i), nil)
		}
	}

	b, err := json.Marshal(r)
	assert.Nil(err)
	assert.NotNil(b)

	assert.NotEmpty(r.String())

	var p []portStatus
	assert.NoError(json.Unmarshal(b, &p))

	for i, v := range p {
		if i%2 == 0 {
			assert.Equal(v.State, PortInUse)
			assert.NotEmpty(v.Bricks)
		}
	}
}
