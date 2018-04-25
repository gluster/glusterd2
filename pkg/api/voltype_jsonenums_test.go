package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolTypeMarshalJSON(t *testing.T) {
	var r VolType

	resp, err := r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Distribute\"")
	assert.Nil(t, err)

	r = 1
	resp, err = r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Replicate\"")
	assert.Nil(t, err)

}

func TestVolTypeUnmarshalJSON(t *testing.T) {
	var r VolType

	err := r.UnmarshalJSON([]byte("\"Distribute\""))
	assert.Nil(t, err)

	err = r.UnmarshalJSON([]byte("\"Replicate\""))
	assert.Nil(t, err)

}
