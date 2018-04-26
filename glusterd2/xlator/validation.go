package xlator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"

	log "github.com/sirupsen/logrus"
)

// ValidationFunc is a function that is invoked during volume set. Each plugin
// or xlator can provide such validation function.
type ValidationFunc func(*volume.Volinfo, string, string) error

func validateReplica(v *volume.Volinfo, key string, value string) error {
	switch key {
	case "metadata-self-heal":
		if v.Subvols[0].ReplicaCount == 1 {
			return errors.New("Option cannot be set for a non replicate volume")
		}
	}
	return nil
}

func validateDht(v *volume.Volinfo, key string, value string) error {
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

// RegisterValidationFunc registers a xlator's validation function for calling
// later during volume set operation.
func RegisterValidationFunc(xlator string, vf ValidationFunc) error {
	xl, err := Find(xlator)
	if err != nil {
		log.WithError(err).WithField("xlator",
			xlator).Error("Could not register xlator validation function")
		return err
	}
	xl.Validate = vf
	return nil
}

func registerAllValidations() error {
	if err := RegisterValidationFunc("afr", validateReplica); err != nil {
		return err
	}
	if err := RegisterValidationFunc("dht", validateDht); err != nil {
		return err
	}
	return nil
}
