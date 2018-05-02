package volgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	var err ErrOptsNotFound = "test"
	e := err.Error()
	assert.Contains(t, e, "test")
}
