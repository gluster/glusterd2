package quota

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

const (
	quotaEnabledKey = "features.quota"
)

func validateOptions(v *volume.Volinfo, key string, value string) error {

	if isQuotaEnabled(v) == false {
		err := fmt.Errorf("Quota not enabled to set this value: '%s'", key)
		return err
	}

	switch key {
	case "deem-statfs":
		return nil
	case "hard-timeout":
		return nil
	case "soft-timeout":
		return nil
	case "alert-time":
		return nil
	case "default-soft-limit":
		return nil
	default:
		err := fmt.Errorf("'%s' is not a valid quota option", key)
		return err
	}
}

func isQuotaEnabled(v *volume.Volinfo) bool {
	val, exists := v.Options[quotaEnabledKey]
	if exists && val == "on" {
		return true
	}
	return false
}

func init() {
	xlator.RegisterValidationFunc(name, validateOptions)
}
