package brickmux

import (
	"github.com/gluster/glusterd2/glusterd2/options"
)

const (
	brickMuxOpKey = "cluster.brick-multiplex"
)

// Enabled returns true if brick multiplexing has been enabled and returns
// falls otherwise.
func Enabled() (bool, error) {

	value, err := options.GetClusterOption(brickMuxOpKey)
	if err != nil {
		return false, err
	}

	ok, err := options.StringToBoolean(value)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// validateOption validates brick mux options
func validateOption(option, value string) error {
	return nil
}

func init() {
	options.RegisterClusterOpValidationFunc(brickMuxOpKey, validateOption)
}
