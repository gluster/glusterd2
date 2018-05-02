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

func TestSignInOut(t *testing.T) {
	inReq := &SignInReq{
		Brick: "/tmp/brick1",
		Port:  1000,
	}
	inRes := &SignInRsp{}
	gfPortMap := NewGfPortmap()
	err := gfPortMap.SignIn(inReq, inRes)
	assert.Nil(t, err)

	OutReq := &SignOutReq{
		Brick: "/tmp/brick1",
		Port:  1000,
	}
	OutRes := &SignOutRsp{}

	err = gfPortMap.SignOut(OutReq, OutRes)
	assert.Nil(t, err)

}
