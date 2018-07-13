package dht

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

var names = [...]string{"distribute", "dht"}

func validateOptions(v *volume.Volinfo, key, value string) error {
	if strings.Contains(key, "readdirplus-for-dir") {
		if value == "on" {
			if v, ok := v.Options["features.cache-invalidation"]; ok && v == "on" {
				return nil
			}
			return fmt.Errorf("Enable \"features.cache-invalidation\" before enabling %s", key)
		}
	}
	return nil
}

func init() {
	for _, name := range names {
		xlator.RegisterValidationFunc(name, validateOptions)
	}
}
