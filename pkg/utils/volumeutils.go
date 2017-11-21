package utils

import (
	"path"

	config "github.com/spf13/viper"
)

// GetVolumeDir returns path to volume directory
func GetVolumeDir(volumeName string) string {
	return path.Join(config.GetString("localstatedir"), "vols", volumeName)
}
