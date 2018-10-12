package options

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/errors"

	"strconv"
)

const (
	clusterOptionsKey string = "clusteroptions"
)

// ClusterOption reperesents a single cluster wide option
type ClusterOption struct {
	Key          string
	DefaultValue string
	Type         OptionType
}

// ClusterOptMap contains list of supported cluster-wide options, default values and value types
var ClusterOptMap = map[string]ClusterOption{
	"cluster.server-quorum-ratio":    {"cluster.server-quorum-ratio", "51", OptionTypePercent},
	"cluster.shared-storage":         {"cluster.shared-storage", "disable", OptionTypeBool},
	"cluster.op-version":             {"cluster.op-version", strconv.Itoa(gdctx.OpVersion), OptionTypeInt},
	"cluster.max-op-version":         {"cluster.max-op-version", strconv.Itoa(gdctx.OpVersion), OptionTypeInt},
	"cluster.brick-multiplex":        {"cluster.brick-multiplex", "disable", OptionTypeBool},
	"cluster.max-bricks-per-process": {"cluster.max-bricks-per-process", "0", OptionTypeInt},
	"cluster.localtime-logging":      {"cluster.localtime-logging", "disable", OptionTypeBool},
}

// ClusterOptions contains cluster-wide attributes
type ClusterOptions struct {
	Options map[string]string
}

// GetClusterOptions gets cluster options from store.
func GetClusterOptions() (*ClusterOptions, error) {
	resp, err := store.Get(context.TODO(), clusterOptionsKey)
	if err != nil {
		return nil, err
	}

	if resp.Count != 1 {
		return nil, errors.ErrClusterOptionsNotFound
	}

	var c ClusterOptions
	if err = json.Unmarshal(resp.Kvs[0].Value, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// GetClusterOption returns the value set for the cluster option specified. If
// the value is not set for the key, it returns the default value for the
// option.
func GetClusterOption(key string) (string, error) {
	globalopt, found := ClusterOptMap[key]
	if !found {
		return "", errors.ErrInvalidClusterOption
	}

	c, err := GetClusterOptions()
	if err != nil && err != errors.ErrClusterOptionsNotFound {
		return "", err
	}

	result := globalopt.DefaultValue
	if value, ok := c.Options[key]; ok {
		result = value
	}

	return result, nil
}

// UpdateClusterOptions stores cluster options in store.
func UpdateClusterOptions(c *ClusterOptions) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	_, err = store.Put(context.TODO(), clusterOptionsKey, string(b))
	return err
}
