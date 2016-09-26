package utils

import (
	"fmt"
	"path"
	"strings"

	config "github.com/spf13/viper"
)

func GetVolumeDir(volumeName string) string {
	return path.Join(config.GetString("localstatedir"), "vols", volumeName)
}

func GetBrickVolFilePath(volumeName string, brickHostName string, brickPath string) string {
	volumeDir := GetVolumeDir(volumeName)
	brickPathWithoutSlashes := strings.Replace(brickPath, "/", "-", -1)
	return fmt.Sprintf("%s/%s.%s.%s.vol", volumeDir, volumeName, brickHostName, brickPathWithoutSlashes)
}
