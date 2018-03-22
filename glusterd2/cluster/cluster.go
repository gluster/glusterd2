package cluster

import (
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"

	"strconv"
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

// IsBrickMuxEnabled returns whether brick multiplexing is enabled
func IsBrickMuxEnabled() (bool, error) {
	val, err := GetGlobalOptionVal("cluster.brick-multiplex")
	if err != nil {
		return false, err
	}

	boolval, err := options.StringToBoolean(val)
	if err != nil {
		return false, err
	}

	return boolval, nil
}

// MaxBricksPerGlusterfsd returns the maximum number of bricks allowed per brick process
func MaxBricksPerGlusterfsd() (int, error) {
	val, err := GetGlobalOptionVal("cluster.max-bricks-per-process")
	if err != nil {
		return 1, err
	}

	limit, err := strconv.Atoi(val)
	if err != nil {
		return 1, err
	}
	return limit, nil
}
