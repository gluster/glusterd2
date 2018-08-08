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
	switch key {
	case "scrub-throttle":
		acceptedThrottleValues := []string{"lazy", "normal", "aggressive"}
		if contains(value, acceptedThrottleValues) {
			return nil
		}
		return fmt.Errorf(
			"invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedThrottleValues, ", "))
	case "scrub-freq":
		acceptedFrequencyValues := []string{"hourly", "daily", "weekly", "biweekly", "monthly"}
		if contains(value, acceptedFrequencyValues) {
			return nil
		}
		return fmt.Errorf(
			"Invalid value specified for option '%s'. Possible values: {%s}",
			key, strings.Join(acceptedFrequencyValues, ", "))
	case "scrub-state":
		acceptedScrubStateValues := []string{"pause", "resume"}
		if contains(value, acceptedScrubStateValues) {
			return nil
		}
		return fmt.Errorf(
			"invalid value specified for option '%s'. possible values: {%s}",
			key, strings.Join(acceptedScrubStateValues, ", "))

	}
	return nil
}

func isBitrotEnabled(v *volume.Volinfo) bool {
	if v, ok := v.Options[keyFeaturesBitrot]; ok && v == "on" {
		return true
	}
	return false
}

func init() {
	xlator.RegisterValidationFunc(name, validateOptions)
}
