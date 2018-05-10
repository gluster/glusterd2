package volgen

import (
	"testing"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsClusterGraph(t *testing.T) {
	resp := isClusterGraph("vol.graph")
	assert.False(t, resp)

	resp = isClusterGraph("cluster.graph")
	assert.True(t, resp)
}

func TestNewClusterGraph(r *testing.T) {
	nodes := make([]*Node, 1)
	var a = qArgs{
		vol: &volume.Volinfo{},
		g:   &Graph{},
		t:   &Node{Children: nodes},
	}

	//nodes having children
	_, err := newClusterGraph(a)
	assert.NotNil(r, err)

	a.vol.Type = 5
	a.t.Children = nil
	//invalid vol type
	_, err = newClusterGraph(a)
	assert.NotNil(r, err)

	a.vol.Type = 0
	a.g.id = "test"
	_, err = newClusterGraph(a)
	//incorect number of bricks
	assert.NotNil(r, err)

	a.t.Voltype = "cluster/replicate"
	_, err = newClusterGraph(a)
	//incorect number of bricks
	assert.NotNil(r, err)

}

func TestGetChildCount(t *testing.T) {
	var vol volume.Volinfo
	vol.Subvols = make([]volume.Subvol, 0)
	vol.Subvols = append(vol.Subvols, volume.Subvol{
		ReplicaCount:  3,
		DisperseCount: 1,
	})
	vol.DistCount = 1

	v := getChildCount("cluster/afr", &vol)
	assert.Equal(t, v, 3)

	v = getChildCount("cluster/dht", &vol)
	assert.Equal(t, v, 1)

	v = getChildCount("cluster/replicate", &vol)
	assert.Equal(t, v, 3)

	v = getChildCount("cluster/disperse", &vol)
	assert.Equal(t, v, 1)

	v = getChildCount("cluster/distribute", &vol)
	assert.Equal(t, v, 1)

	v = getChildCount("", &vol)
	assert.Equal(t, v, 0)
}

func TestNewClientNodes(t *testing.T) {
	var (
		vol   volume.Volinfo
		extra map[string]string
	)
	vol.Subvols = make([]volume.Subvol, 0)
	vol.Subvols = append(vol.Subvols, volume.Subvol{
		ReplicaCount:  3,
		DisperseCount: 1,
	})

	nodes, err := newClientNodes(&vol, "cluster", extra)
	assert.Empty(t, nodes)
	assert.Nil(t, err)

	vol.Subvols[0].Bricks = make([]brick.Brickinfo, 1)
	vol.Subvols[0].Bricks = append(vol.Subvols[0].Bricks, brick.Brickinfo{
		ID:   uuid.NewRandom(),
		Path: "temp/bricks",
	})

	nodes, err = newClientNodes(&vol, "cluster", extra)
	assert.Empty(t, nodes)
	assert.Contains(t, err.Error(), "not found")

}
