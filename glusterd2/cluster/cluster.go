package cluster

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"

	"strconv"
)

const (
	// ClusterPrefix represents the etcd end-point for cluster wide attributes
	ClusterPrefix string = "cluster/"
)

// GlobalOptMap contains list of supported cluster-wide options, default values and value types
var GlobalOptMap = map[string]GlobalOption{
	"cluster.server-quorum-ratio":    {"cluster.server-quorum-ratio", "51", options.OptionTypePercent},
	"cluster.shared-storage":         {"cluster.shared-storage", "disable", options.OptionTypeBool},
	"cluster.op-version":             {"cluster.op-version", strconv.Itoa(gdctx.OpVersion), options.OptionTypeInt},
	"cluster.max-op-version":         {"cluster.max-op-version", strconv.Itoa(gdctx.OpVersion), options.OptionTypeInt},
	"cluster.brick-multiplex":        {"cluster.brick-multiplex", "disable", options.OptionTypeBool},
	"cluster.max-bricks-per-process": {"cluster.max-bricks-per-process", "0", options.OptionTypeInt},
	"cluster.localtime-logging":      {"cluster.localtime-logging", "disable", options.OptionTypeBool},
}

// Cluster contains cluster-wide attributes
type Cluster struct {
	Options map[string]string
}

// GlobalOption reperesents cluster wide options
type GlobalOption struct {
	Key          string
	DefaultValue string
	Type         options.OptionType
}

// LoadClusterAttributes updates etcd with cluster attributes if not already present
func LoadClusterAttributes() error {
	var clstr Cluster
	resp, err := store.Store.Get(context.TODO(), ClusterPrefix)
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		// If cluster instance isn't available add it to etcd
		clstr.Options = make(map[string]string)
		for k := range GlobalOptMap {
			clstr.Options[k] = GlobalOptMap[k].DefaultValue
		}

		data, _ := json.Marshal(clstr)
		if _, err := store.Store.Put(context.TODO(), ClusterPrefix, string(data)); err != nil {
			return err
		}
	}

	// If cluster instance is already available on etcd just return
	return nil
}
