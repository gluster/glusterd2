package dht

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

var names = [...]string{"distribute", "dht"}

func validateOptions(v *volume.Volinfo, key string, value string) error {
	var err error
	if strings.Contains(key, "readdirplus-for-dir") {
		if value == "on" {
			val, exists := v.Options["features.cache-invalidation"]
			if exists && val == "on" {
				return nil
			}
			err = fmt.Errorf("Enable \"features.cache-invalidation\" before enabling %s",
				key)
			return err
		}
	}
	return nil
}

func init() {
	for _, name := range names {
		xlator.RegisterValidationFunc(name, validateOptions)
	}
}
