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
	resp, err := store.Store.Get(context.TODO(), ClusterPrefix)
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
	if _, err := store.Store.Put(context.TODO(), ClusterPrefix, string(data)); err != nil {
		return err
	}
	return nil
}

// GetGlobalOptionVal returns global option value
func GetGlobalOptionVal(key string) (string, error) {
	globalopt, found := GlobalOptMap[key]
	if !found {
		return "", errors.ErrInvalidGlobalOption
	}

	c, err := GetCluster()
	// ErrClusterNotFound here implies that no global option has yet been explicitly set. Ignoring it.
	if err != nil && err != errors.ErrClusterNotFound {
		return "", err
	}

	var val string
	if c == nil {
		val = globalopt.DefaultValue
	} else {
		var found bool
		val, found = c.Options[key]
		if !found {
			val = globalopt.DefaultValue
		}
	}

	return val, nil
}
