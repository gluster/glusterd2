package brickmux

import (
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/options"
	"github.com/gluster/glusterd2/pkg/errors"
)

const (
	brickMuxOpKey               = "cluster.brick-multiplex"
	brickMuxMaxBricksPerProcKey = "cluster.max-bricks-per-process"
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

func getMaxBricksPerProcess() (int, error) {
	value, err := options.GetClusterOption(brickMuxMaxBricksPerProcKey)
	if err != nil {
		return -1, err
	}

	maxBricksPerProcess, err := strconv.Atoi(value)
	if err != nil {
		return -1, err
	}

	return maxBricksPerProcess, nil
}

// validateOption validates brick mux options
func validateOption(option, value string) error {
	if option == "cluster.max-bricks-per-process" {
		_, err := strconv.Atoi(value)
		if err != nil {
			return errors.ErrInvalidIntValue
		}
	}

	return nil
}

func init() {
	options.RegisterClusterOpValidationFunc(brickMuxOpKey, validateOption)
	options.RegisterClusterOpValidationFunc(brickMuxMaxBricksPerProcKey, validateOption)
}
