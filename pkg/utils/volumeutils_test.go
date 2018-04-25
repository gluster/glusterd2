package utils

import (
	"testing"

	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetVolumeDir(t *testing.T) {
	config.Set("localstatedir", "/var/lib/glusterd2")

	resp := GetVolumeDir("testvolume")
	assert.Equal(t, resp, "/var/lib/glusterd2/vols/testvolume")

}

func TestGetSnapshotDir(t *testing.T) {
	config.Set("localstatedir", "/var/lib/glusterd2")

	resp := GetSnapshotDir("testvolume")
	assert.Equal(t, resp, "/var/lib/glusterd2/snaps/testvolume")

}
