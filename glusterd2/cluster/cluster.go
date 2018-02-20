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
