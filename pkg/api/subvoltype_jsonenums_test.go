package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubvolMarshalJSON(t *testing.T) {
	var r SubvolType

	resp, err := r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Distribute\"")
	assert.Nil(t, err)

	r = 1
	resp, err = r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Replicate\"")
	assert.Nil(t, err)

	r = 3
	resp, err = r.MarshalJSON()
	assert.Equal(t, len(resp), 0)
	assert.Contains(t, err.Error(), "invalid SubvolType")

}

func TestSubVolUnmarshalJSON(t *testing.T) {
	var r SubvolType

	err := r.UnmarshalJSON([]byte("\"Distribute\""))
	assert.Nil(t, err)

	err = r.UnmarshalJSON([]byte("\"Replicate\""))
	assert.Nil(t, err)

	err = r.UnmarshalJSON([]byte("\"test\""))
	assert.Contains(t, err.Error(), "invalid SubvolType")

	err = r.UnmarshalJSON([]byte("1"))
	assert.Contains(t, err.Error(), "SubvolType should be a string")

}
