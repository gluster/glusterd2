package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHumanReadable checks if the function returns the right size with unit
// for the input given in MB
func TestHumanReadable(t *testing.T) {
	assert.Equal(t, "20.0 MB", humanReadable(20))
	assert.Equal(t, "1.0 GB", humanReadable(1024))
	assert.Equal(t, "1.5 GB", humanReadable(1536))
	assert.Equal(t, "1.0 TB", humanReadable(1048576))
}
