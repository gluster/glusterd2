package bitrot

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

func contains(s string, list []string) bool {
	for _, val := range list {
		if s == val {
			return true
		}
	}
	return false
}

func validateOptions(v *volume.Volinfo, key string, value string) error {
	var err error
	switch key {
	case "scrub-throttle":
		acceptedThrottleValues := []string{"lazy", "normal", "aggressive"}
		if contains(value, acceptedThrottleValues) {
			return nil
		}
		err = fmt.Errorf("Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedThrottleValues, ", "))
		return err
	case "scrub-freq":
		acceptedFrequencyValues := []string{"hourly", "daily", "weekly", "biweekly", "monthly"}
		if contains(value, acceptedFrequencyValues) {
			return nil
		}
		err = fmt.Errorf("Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedFrequencyValues, ", "))
		return err
	case "scrub-state":
		acceptedScrubStateValues := []string{"pause", "resume"}
		if contains(value, acceptedScrubStateValues) {
			return nil
		}
		err = fmt.Errorf("Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedScrubStateValues, ", "))
		return err
	}
	return nil
}

func isBitrotEnabled(v *volume.Volinfo) bool {
	val, exists := v.Options[keyFeaturesBitrot]
	if exists && val == "on" {
		return true
	}
	return false
}

func init() {
	xlator.RegisterValidationFunc(name, validateOptions)
}
