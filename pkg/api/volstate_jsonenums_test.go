package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolStateMarshalJSON(t *testing.T) {
	var r VolState

	resp, err := r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Created\"")
	assert.Nil(t, err)

	r = 1
	resp, err = r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Started\"")
	assert.Nil(t, err)

}

func TestVolStateUnmarshalJSON(t *testing.T) {
	var r VolState

	err := r.UnmarshalJSON([]byte("\"Created\""))
	assert.Nil(t, err)

	err = r.UnmarshalJSON([]byte("\"Started\""))
	assert.Nil(t, err)

}
