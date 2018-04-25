package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrickMarshalJSON(t *testing.T) {
	var r BrickType

	resp, err := r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Brick\"")
	assert.Nil(t, err)

	r = 1
	resp, err = r.MarshalJSON()
	assert.Equal(t, string(resp), "\"Arbiter\"")
	assert.Nil(t, err)

	r = 2
	resp, err = r.MarshalJSON()
	fmt.Println(resp, err)
	assert.Equal(t, len(resp), 0)
	assert.Contains(t, err.Error(), "invalid BrickType")

}

func TestBrickUnmarshalJSON(t *testing.T) {
	var r BrickType

	err := r.UnmarshalJSON([]byte("\"Brick\""))
	assert.Nil(t, err)

	err = r.UnmarshalJSON([]byte("\"Arbiter\""))
	assert.Nil(t, err)

	err = r.UnmarshalJSON([]byte("\"test\""))
	assert.Contains(t, err.Error(), "invalid BrickType")

	err = r.UnmarshalJSON([]byte("1"))
	assert.Contains(t, err.Error(), "BrickType should be a string")

}
