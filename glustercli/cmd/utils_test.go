package cmd

import (
	"testing"

	gutils "github.com/gluster/glusterd2/pkg/utils"

	"github.com/stretchr/testify/assert"
)

// TestHumanReadable checks if the function returns the right size with unit
// for the input given in bytes
func TestHumanReadable(t *testing.T) {
	assert.Equal(t, "900.0 B", humanReadable(900))
	assert.Equal(t, "1.0 KiB", humanReadable(1024))
	assert.Equal(t, "1.5 KiB", humanReadable(1536))
	assert.Equal(t, "1.0 MiB", humanReadable(1*gutils.MiB))
	assert.Equal(t, "1.5 MiB", humanReadable(1.5*gutils.MiB))
	assert.Equal(t, "20.0 KiB", humanReadable(20*gutils.KiB))
	assert.Equal(t, "20.0 MiB", humanReadable(20*gutils.MiB))
	assert.Equal(t, "1.0 GiB", humanReadable(1*gutils.GiB))
	assert.Equal(t, "1.5 GiB", humanReadable(1536*gutils.MiB))
	assert.Equal(t, "1.0 TiB", humanReadable(1*gutils.TiB))
}
