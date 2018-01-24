package xlator

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/volume"
)

type validationFunc func(*volume.Volinfo, string, string) error

// Sample validation function
func validateReplica(v *volume.Volinfo, key string, value string) error {
	switch key {
	case "metadata-self-heal":
		if v.Subvols[0].ReplicaCount == 1 {
			return errors.New("Option cannot be set for a non replicate volume")
		}
	}
	return nil
}

func registerValidation(xlator string, vf validationFunc) error {
	xl, err := Find(xlator)
	if err != nil {
		return err
	}
	xl.Validate = vf
	return nil
}

func registerAllValidations() error {
	if err := registerValidation("afr", validateReplica); err != nil {
		return err
	}
	return nil
}
