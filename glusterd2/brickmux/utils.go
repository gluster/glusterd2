package brickmux

import (
	"os"
	"path"

	config "github.com/spf13/viper"
)

// getBrickVolfilePath returns path correspponding to the volfileID. Since, brick volfiles are now stored on
// disk, added a check to ensure the volfile exists on the path.
func getBrickVolfilePath(volfileID string) (string, error) {

	volfilePath := path.Join(config.GetString("localstatedir"),
		"volfiles", volfileID+".vol")

	if _, err := os.Stat(volfilePath); os.IsNotExist(err) {
		return "", err
	}

	return volfilePath, nil
}
