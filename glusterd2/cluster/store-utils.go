package cluster

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/errors"
)

const (
	// ClusterPrefix represents the etcd end-point for cluster wide attributes
	ClusterPrefix string = "cluster/"
)

// GetCluster gets cluster object instace from store
func GetCluster() (*Cluster, error) {
	resp, err := store.Get(context.TODO(), ClusterPrefix)
	if err != nil {
		return nil, err
	}

	if resp.Count != 1 {
		return nil, errors.ErrClusterNotFound
	}

	var c Cluster
	if err = json.Unmarshal(resp.Kvs[0].Value, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdateCluster updates cluster instance in etcd store
func UpdateCluster(c *Cluster) error {
	data, _ := json.Marshal(c)
	_, err := store.Put(context.TODO(), ClusterPrefix, string(data))
	return err
}
