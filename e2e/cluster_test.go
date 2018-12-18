package e2e

import (
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

func testGetClusterOptions(t *testing.T) {
	r := require.New(t)
	clusterOps, err := client.GetClusterOption()
	r.Nil(err)
	r.NotNil(clusterOps)
}

func ClusterOptionsSet(t *testing.T) {
	r := require.New(t)
	optReq := api.ClusterOptionReq{
		Options: map[string]string{"cluster.brick-multiplex": "on"},
	}
	err := client.ClusterOptionSet(optReq)
	r.Nil(err)
	clusterOps, err := client.GetClusterOption()
	for _, ops := range clusterOps {
		if ops.Key == "cluster.brick-multiplex" {
			r.Equal(ops.Value, "on")
		}
	}
}

// TestClusterOption creates a cluster
func TestClusterOption(t *testing.T) {
	var err error

	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml", "./config/3.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err = initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	t.Run("Get Cluster Options", testGetClusterOptions)
}
