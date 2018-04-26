package afr

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

var names = [...]string{"replicate", "afr"}

func validateOptions(v *volume.Volinfo, key string, value string) error {
	switch key {
	case "metadata-self-heal":
		if v.Subvols[0].ReplicaCount == 1 {
			return errors.New("Option cannot be set for a non replicate volume")
		}
	}
	return nil
}

func init() {
	for _, name := range names {
		xlator.RegisterValidationFunc(name, validateOptions)
	}
}
