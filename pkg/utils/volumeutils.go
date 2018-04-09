package utils

import (
	"path"

	config "github.com/spf13/viper"
)

// GetVolumeDir returns path to volume directory
func GetVolumeDir(volumeName string) string {
	return path.Join(config.GetString("localstatedir"), "vols", volumeName)
}

// GetSnapshotDir returns path to snapshot directory
func GetSnapshotDir(snapName string) string {
	return path.Join(config.GetString("localstatedir"), "snaps", snapName)
}
