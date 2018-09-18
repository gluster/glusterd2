package pmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGfPortmap(t *testing.T) {
	gfPortMap := NewGfPortmap()
	assert.NotNil(t, gfPortMap)

	name := gfPortMap.Name()
	assert.Equal(t, name, "Gluster Portmap")

	num := gfPortMap.Number()
	assert.Equal(t, num, uint32(portmapProgNum))

	ver := gfPortMap.Version()
	assert.Equal(t, ver, uint32(portmapProgVersion))

	pro := gfPortMap.Version()
	assert.NotNil(t, pro, portmapProgVersion)

}

func TestPortByBrick(t *testing.T) {
	gfPortMap := NewGfPortmap()
	assert.NotNil(t, gfPortMap)

	var req PortByBrickReq
	var res PortByBrickRsp
	err := gfPortMap.PortByBrick(&req, &res)
	assert.Nil(t, err)
}
