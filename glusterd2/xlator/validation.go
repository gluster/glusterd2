package xlator

import (
	"errors"
	"fmt"
	"strings"

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

func validateBitrot(v *volume.Volinfo, key string, value string) error {
	var err error
	switch key {
	case "scrub-throttle":
		acceptedThrottleValues := []string{"lazy", "normal", "aggressive"}
		if Contains(value, acceptedThrottleValues) {
			return nil
		}
		err = fmt.Errorf("Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedThrottleValues, ", "))
		return err
	case "scrub-freq":
		acceptedFrequencyValues := []string{"hourly", "daily", "weekly", "biweekly", "monthly"}
		if Contains(value, acceptedFrequencyValues) {
			return nil
		}
		err = fmt.Errorf("Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedFrequencyValues, ", "))
		return err
	case "scrub-state":
		acceptedScrubStateValues := []string{"pause", "resume"}
		if Contains(value, acceptedScrubStateValues) {
			return nil
		}
		err = fmt.Errorf("Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedScrubStateValues, ", "))
		return err
	}
	return nil
}

func validateQuota(v *volume.Volinfo, key string, value string) error {

	// Check if quota is already enabled
	if volume.IsQuotaEnabled(v) == false {
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
	if err := registerValidation("bit-rot", validateBitrot); err != nil {
		return err
	}
	if err := registerValidation("quota", validateQuota); err != nil {
		return err
	}
	return nil
}
